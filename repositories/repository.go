package repositories

import (
	"fmt"
	"io"
	"jc-financas/models"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var storeSessions = sessions.NewCookieStore([]byte("xPjrXZsDfdlwlYzFcWZQZ92f6x9IuTkHp_m7KZTlPlg=")) // Defina uma chave secreta
var DB *gorm.DB

func CreateUser(model *models.User) error {
	return DB.Create(model).Error
}

// CRUD Handler Genérico
type CRUDHandler struct {
	DB        *gorm.DB
	Model     interface{}
	TableName string
}

func (h *CRUDHandler) Register(c echo.Context) error {
	var user models.User

	// Bind para preencher o struct com os dados do formulário
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

	// Validações simples
	if user.Name == "" || user.Email == "" || user.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Nome, email e senha são obrigatórios"})
	}

	// Encriptar a senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao encriptar senha"})
	}

	// Substituir a senha em texto puro pela senha encriptada
	user.Password = string(hashedPassword)
	user.Role = "admin"
	// Salvar no banco de dados
	if err := h.DB.Create(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao registrar usuário",
			"message": err.Error(),
		})
	}

	defaultTeam := models.Team{
		Name:   user.Name,
		UserID: user.ID,
	}

	if err := h.DB.Create(&defaultTeam).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro criar o time padrao",
			"message": err.Error(),
		})
	}

	// Atualizar o team_id do usuário logado
	if err := h.DB.Model(&models.User{}).Where("id = ?", user.ID).Update("team_id", defaultTeam.ID).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao associar o time ao usuário"})
	}

	user.TeamID = defaultTeam.ID

	if err := salvaSessao(c, user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao salvar a sessao",
			"message": err.Error(),
		})
	}
	// return c.JSON(http.StatusCreated, user)
	return c.Redirect(http.StatusSeeOther, "/home")
}

func salvaSessao(c echo.Context, user models.User) error {
	session, _ := storeSessions.Get(c.Request(), "session-id")
	session.Values["user"] = user
	session.Save(c.Request(), c.Response())

	err := session.Save(c.Request(), c.Response())
	if err != nil {
		return err
	}
	return nil
}

func (h *CRUDHandler) Create(c echo.Context) error {
	// Criar a instância do modelo dinamicamente
	record := reflect.New(reflect.TypeOf(h.Model).Elem()).Interface()

	// Fazer o bind dos dados da requisição na struct
	if err := c.Bind(record); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos", "message": err.Error()})
	}

	// Converter a data corretamente
	dateStr := c.FormValue("date_at")
	fmt.Println(dateStr)
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	fmt.Println(parsedDate)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Data inválida"})
	}
	reflect.ValueOf(record).Elem().FieldByName("DateAt").Set(reflect.ValueOf(parsedDate))

	// Salvar no banco de dados
	if err := h.DB.Create(record).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar"})
	}

	return c.JSON(http.StatusCreated, record)
}

// Listar Registros
func (h *CRUDHandler) List(c echo.Context) error {
	// Criar um slice do tipo correto dinamicamente
	records := reflect.New(reflect.SliceOf(reflect.TypeOf(h.Model).Elem())).Interface()

	// Buscar os registros no banco
	if err := h.DB.Table(h.TableName).Find(records).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar registros"})
	}

	return c.JSON(http.StatusOK, records)
}

// Obter Registro por ID
func (h *CRUDHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	record := h.Model

	if err := h.DB.First(&record, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Registro não encontrado"})
	}

	return c.JSON(http.StatusOK, record)
}

// Atualizar Registro
func (h *CRUDHandler) Update(c echo.Context) error {
	id := c.Param("id")
	updatedRecord := h.Model

	if err := h.DB.First(&updatedRecord, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Registro não encontrado"})
	}

	if err := c.Bind(&updatedRecord); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

	if err := h.DB.Save(&updatedRecord).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao atualizar"})
	}

	return c.JSON(http.StatusOK, updatedRecord)
}

// Deletar Registro
func (h *CRUDHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	record := h.Model

	if err := h.DB.Delete(&record, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao deletar"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Registro deletado com sucesso"})
}

func (h *CRUDHandler) CreateCategory(c echo.Context) error {
	// Obter a sessão e o usuário logado
	session, err := storeSessions.Get(c.Request(), "session-id")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Sessão inválida"})
	}
	user, ok := session.Values["user"].(models.User)
	if !ok || user.ID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Usuário não autenticado"})
	}

	// Criar a instância de Transaction
	model := models.Category{}
	if err := c.Bind(&model); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

	// Validações simples
	if model.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "O nome da conta é obrigatório"})
	}

	// Definir TeamID
	model.TeamID = user.TeamID

	// Salvar o registro no banco
	if err := h.DB.Create(&model).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar"})
	}

	return c.JSON(http.StatusCreated, model)
}

// Listar transactions do usuário logado
func (h *CRUDHandler) ListCategories(c echo.Context) error {
	// Obter a sessão e o usuário logado
	session, err := storeSessions.Get(c.Request(), "session-id")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Sessão inválida"})
	}

	user, ok := session.Values["user"].(models.User)
	if !ok || user.ID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Usuário não autenticado"})
	}

	// Criar um slice
	var records []models.Category

	// Buscar apenas as transactions do TeamID do usuário logado
	if err := h.DB.Where("team_id = ?", user.TeamID).
		Order("id DESC").
		Find(&records).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar registros"})
	}

	return c.JSON(http.StatusOK, records)
}

func (h *CRUDHandler) CreateAccount(c echo.Context) error {
	// Obter a sessão e o usuário logado
	session, err := storeSessions.Get(c.Request(), "session-id")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Sessão inválida"})
	}
	user, ok := session.Values["user"].(models.User)
	if !ok || user.ID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Usuário não autenticado"})
	}

	// Criar a instância de Transaction
	model := models.Account{}
	if err := c.Bind(&model); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

	// Definir TeamID
	model.TeamID = user.TeamID

	// Salvar o registro no banco
	if err := h.DB.Create(&model).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar"})
	}

	return c.JSON(http.StatusCreated, model)
}

// Listar transactions do usuário logado
func (h *CRUDHandler) ListAccounts(c echo.Context) error {
	// Obter a sessão e o usuário logado
	session, err := storeSessions.Get(c.Request(), "session-id")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Sessão inválida"})
	}

	user, ok := session.Values["user"].(models.User)
	if !ok || user.ID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Usuário não autenticado"})
	}

	// Criar um slice
	var records []models.Account

	// Buscar apenas as transactions do TeamID do usuário logado
	if err := h.DB.Where("team_id = ?", user.TeamID).
		Order("id DESC").
		Find(&records).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar registros"})
	}

	return c.JSON(http.StatusOK, records)
}

func (h *CRUDHandler) CreateTransaction(c echo.Context) error {
	// Obter a sessão e o usuário logado
	session, err := storeSessions.Get(c.Request(), "session-id")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Sessão inválida"})
	}
	user, ok := session.Values["user"].(models.User)
	if !ok || user.ID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Usuário não autenticado"})
	}

	// Criar a instância de Transaction
	transaction := models.Transaction{}
	if err := c.Bind(&transaction); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

	// Definir TeamID
	transaction.TeamID = user.TeamID

	// Tratamento para comprovante (upload de arquivo)

	file, err := c.FormFile("proof")
	if err == nil {
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao abrir o arquivo"})
		}
		defer src.Close()

		// Criar o diretório do ano/mês
		year, month, _ := time.Now().Date()
		dir := fmt.Sprintf("static/comprovantes/%d/%02d/", year, month)
		os.MkdirAll(dir, os.ModePerm)

		// Gerar o caminho do arquivo
		filename := fmt.Sprintf("%s%s", uuid.New().String(), filepath.Ext(file.Filename))
		dstPath := filepath.Join(dir, filename)

		// Criar o arquivo no servidor
		dst, err := os.Create(dstPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar o arquivo"})
		}
		defer dst.Close()
		io.Copy(dst, src)

		// Salvar o caminho do comprovante
		transaction.Proof = &dstPath
	}

	// Salvar a transação no banco
	if err := h.DB.Create(&transaction).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar"})
	}

	return c.JSON(http.StatusCreated, transaction)
}

func (h *CRUDHandler) ListTransactions(c echo.Context) error {
	// Obter a sessão e o usuário logado
	session, err := storeSessions.Get(c.Request(), "session-id")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Sessão inválida"})
	}

	user, ok := session.Values["user"].(models.User)
	if !ok || user.ID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Usuário não autenticado"})
	}

	// Parâmetros do DataTables
	start, _ := strconv.Atoi(c.QueryParam("start"))
	length, _ := strconv.Atoi(c.QueryParam("length"))
	search := c.QueryParam("search[value]")
	orderColumn := c.QueryParam("order[0][column]") // Índice da coluna
	orderDir := c.QueryParam("order[0][dir]")       // Direção (asc ou desc)

	// Definir ordenação padrão
	orderBy := "date DESC"
	if orderColumn != "" && orderDir != "" {
		columns := []string{"id", "date", "amount", "category_id", "account_id"} // Defina as colunas do DataTables
		colIndex, err := strconv.Atoi(orderColumn)
		if err == nil && colIndex >= 0 && colIndex < len(columns) {
			orderBy = columns[colIndex] + " " + orderDir
		}
	}

	// Criar a query inicial
	query := h.DB.
		Preload("Category").
		Preload("Account").
		Where("team_id = ?", user.TeamID)

	// Aplicar filtro de pesquisa
	if search != "" {
		query = query.Where("description LIKE ? OR amount::text LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Aplicar paginação e ordenação
	var transactions []models.Transaction
	if err := query.Order(orderBy).Offset(start).Limit(length).Find(&transactions).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar registros"})
	}

	// Contar total de registros sem filtros
	var totalRecords int64
	h.DB.Model(&models.Transaction{}).Where("team_id = ?", user.TeamID).Count(&totalRecords)

	// Contar total de registros filtrados
	totalFiltered := totalRecords

	// Retornar resposta no formato esperado pelo DataTables
	response := map[string]interface{}{
		"draw":            c.QueryParam("draw"),
		"recordsTotal":    totalRecords,
		"recordsFiltered": totalFiltered,
		"data":            transactions,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *CRUDHandler) ListUsers(c echo.Context) error {
	// Obter a sessão e o usuário logado
	session, err := storeSessions.Get(c.Request(), "session-id")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Sessão inválida"})
	}

	user, ok := session.Values["user"].(models.User)
	if !ok || user.ID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Usuário não autenticado"})
	}

	// Parâmetros do DataTables
	start, _ := strconv.Atoi(c.QueryParam("start"))
	length, _ := strconv.Atoi(c.QueryParam("length"))
	search := c.QueryParam("search[value]")
	// orderColumn := c.QueryParam("order[0][column]") // Índice da coluna
	// orderDir := c.QueryParam("order[0][dir]")       // Direção (asc ou desc)

	// Definir ordenação padrão
	orderBy := "id DESC"

	// Criar a query inicial
	query := h.DB.
		Where("team_id = ?", user.TeamID)

	// Aplicar filtro de pesquisa
	if search != "" {
		// query = query.Where("description LIKE ? OR amount::text LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Aplicar paginação e ordenação
	var transactions []models.User
	if err := query.Order(orderBy).Offset(start).Limit(length).Find(&transactions).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao buscar registros",
			"message": err.Error(),
		})
	}

	// Contar total de registros sem filtros
	var totalRecords int64
	h.DB.Model(&models.User{}).Where("team_id = ?", user.TeamID).Count(&totalRecords)

	// Contar total de registros filtrados
	totalFiltered := totalRecords

	// Retornar resposta no formato esperado pelo DataTables
	response := map[string]interface{}{
		"draw":            c.QueryParam("draw"),
		"recordsTotal":    totalRecords,
		"recordsFiltered": totalFiltered,
		"data":            transactions,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *CRUDHandler) CreateUser(c echo.Context) error {
	// Obter a sessão e o usuário logado
	session, err := storeSessions.Get(c.Request(), "session-id")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Sessão inválida"})
	}
	user, ok := session.Values["user"].(models.User)
	if !ok || user.ID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Usuário não autenticado"})
	}

	// Criar a instância de Transaction
	model := models.User{}
	if err := c.Bind(&model); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

	// Definir TeamID
	model.TeamID = user.TeamID

	// Salvar a transação no banco
	if err := h.DB.Create(&model).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao salvar",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, model)
}
