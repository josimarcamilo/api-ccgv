package repositories

import (
	"bufio"
	"fmt"
	"io"
	"jc-financas/models"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
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

	var updatedRecord models.Category
	if err := h.DB.First(&updatedRecord, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Registro não encontrado"})
	}

	if err := c.Bind(&updatedRecord); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Dados inválidos",
			"message": err.Error(),
		})
	}

	if err := h.DB.Save(&updatedRecord).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao atualizar"})
	}

	return c.JSON(http.StatusOK, updatedRecord)
}

// Atualizar Registro
func (h *CRUDHandler) UpdateAccount(c echo.Context) error {
	id := c.Param("id")

	var updatedRecord models.Account
	if err := h.DB.First(&updatedRecord, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Registro não encontrado"})
	}

	if err := c.Bind(&updatedRecord); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Dados inválidos",
			"message": err.Error(),
		})
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
	toSelect := c.QueryParam("toselect")
	if toSelect != "" {
		var records []models.Category
		if err := h.DB.
			Where("team_id = ?", user.TeamID).Order("id ASC").Find(&records).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error":   "Erro ao buscar registros para select",
				"message": err.Error(),
			})
		}
		return c.JSON(http.StatusOK, records)
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
	var records []models.Category
	if err := query.Order(orderBy).Offset(start).Limit(length).Find(&records).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao buscar registros",
			"message": err.Error(),
		})
	}

	// Contar total de registros sem filtros
	var totalRecords int64
	h.DB.Model(&models.Category{}).Where("team_id = ?", user.TeamID).Count(&totalRecords)

	// Contar total de registros filtrados
	totalFiltered := totalRecords

	// Retornar resposta no formato esperado pelo DataTables
	response := map[string]interface{}{
		"draw":            c.QueryParam("draw"),
		"recordsTotal":    totalRecords,
		"recordsFiltered": totalFiltered,
		"data":            records,
	}

	return c.JSON(http.StatusOK, response)
}

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

	toSelect := c.QueryParam("toselect")
	if toSelect != "" {
		var records []models.Account
		if err := h.DB.
			Where("team_id = ?", user.TeamID).Order("name ASC").Find(&records).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error":   "Erro ao buscar registros para select",
				"message": err.Error(),
			})
		}
		return c.JSON(http.StatusOK, records)
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
	var records []models.Account
	if err := query.Order(orderBy).Offset(start).Limit(length).Find(&records).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao buscar registros",
			"message": err.Error(),
		})
	}

	// Contar total de registros sem filtros
	var totalRecords int64
	h.DB.Model(&models.Account{}).Where("team_id = ?", user.TeamID).Count(&totalRecords)

	// Contar total de registros filtrados
	totalFiltered := totalRecords

	// Retornar resposta no formato esperado pelo DataTables
	response := map[string]interface{}{
		"draw":            c.QueryParam("draw"),
		"recordsTotal":    totalRecords,
		"recordsFiltered": totalFiltered,
		"data":            records,
	}

	return c.JSON(http.StatusOK, response)
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

	newTransaction := models.Transaction{}
	newTransaction.TeamID = user.TeamID

	typeParam := c.FormValue("date")
	if typeParam == "1" {
		// newTransaction.Type, _ = strconv.Atoi(typeParam)
	}

	if err := c.Bind(&newTransaction); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Dados inválidos",
			"message": err.Error(),
		})
	}

	file, err := c.FormFile("proof")
	if err == nil {
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao abrir o arquivo"})
		}
		defer src.Close()
		dateStr := c.FormValue("date")

		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro converter a data"})
		}

		// Criar o diretório do ano/mês
		anoMes := t.Format("2006/01")
		dir := fmt.Sprintf("static/comprovantes/%v/", anoMes)
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
		newTransaction.Proof = &dstPath
	}

	// Salvar a transação no banco
	if err := h.DB.Create(&newTransaction).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar"})
	}

	accountDestination := c.FormValue("account_destination")

	if accountDestination != "" {
		acDes, _ := strconv.ParseUint(accountDestination, 10, 64)
		var accDestination uint = uint(acDes)

		transactionDestination := models.Transaction{}
		transactionDestination.TeamID = newTransaction.TeamID
		transactionDestination.Date = newTransaction.Date
		transactionDestination.Type = 1
		transactionDestination.Description = newTransaction.Description
		transactionDestination.CategoryID = newTransaction.CategoryID
		transactionDestination.AccountID = accDestination
		transactionDestination.TransactionOrigin = &newTransaction.ID
		transactionDestination.Value = newTransaction.Value

		if err := h.DB.Create(&transactionDestination).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error":   "Erro ao salvar",
				"message": "Erro ao salvar na conta de destino",
				"detail":  err.Error(),
			})
		}

	}

	return c.JSON(http.StatusCreated, newTransaction)
}

type TransactionOFX struct {
	Type        string
	Date        string
	Amount      float64
	FITID       string
	CheckNum    string
	Description string
}

func (h *CRUDHandler) ImportOFX(c echo.Context) error {
	// Obter a sessão e o usuário logado
	session, err := storeSessions.Get(c.Request(), "session-id")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":   "Sessão inválida",
			"message": "Sessão inválida",
		})
	}
	user, ok := session.Values["user"].(models.User)
	if !ok || user.ID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":   "Usuário não autenticado",
			"message": "Usuário não autenticado",
		})
	}

	file, err := c.FormFile("file_ofx")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro",
			"message": "Erro ao ler o arquivo",
		})
	}
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao abrir o arquivo",
			"message": err.Error(),
		})
	}
	defer src.Close()
	scanner := bufio.NewScanner(src)
	transactions := []TransactionOFX{}
	var current TransactionOFX
	rgxDate := regexp.MustCompile(`\d{8}`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "<STMTTRN>") {
			current = TransactionOFX{}
		} else if strings.HasPrefix(line, "<TRNTYPE>") {
			current.Type = strings.TrimPrefix(line, "<TRNTYPE>")
		} else if strings.HasPrefix(line, "<DTPOSTED>") {
			match := rgxDate.FindString(line)
			if match != "" {
				parsedDate, _ := time.Parse("20060102", match)
				current.Date = parsedDate.Format("2006-01-02")
			}
		} else if strings.HasPrefix(line, "<TRNAMT>") {
			amount, _ := strconv.ParseFloat(strings.TrimPrefix(line, "<TRNAMT>"), 64)
			current.Amount = amount
		} else if strings.HasPrefix(line, "<FITID>") {
			current.FITID = strings.TrimPrefix(line, "<FITID>")
		} else if strings.HasPrefix(line, "<CHECKNUM>") {
			current.CheckNum = strings.TrimPrefix(line, "<CHECKNUM>")
		} else if strings.HasPrefix(line, "<MEMO>") {
			current.Description = strings.TrimPrefix(line, "<MEMO>")
		} else if strings.HasPrefix(line, "</STMTTRN>") {
			transactions = append(transactions, current)
		}
	}

	if err := scanner.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao ler o arquivo",
			"message": err.Error(),
		})
	}

	fmt.Printf("%-10s %-12s %-10s %-15s %-10s %s\n", "Tipo", "Data", "Valor", "FITID", "Cheque", "Descrição")
	fmt.Println(strings.Repeat("-", 70))
	for _, tx := range transactions {
		fmt.Printf("%-10s %-12s %-10.2f %-15s %-10s %s\n", tx.Type, tx.Date, tx.Amount, tx.FITID, tx.CheckNum, tx.Description)
		record := models.Transaction{}
		if err := h.DB.Where("team_id = ?", user.TeamID).Where("external_id = ?", tx.FITID).First(&record).Error; err != nil {
			//criar
			record.TeamID = user.TeamID
			record.Date = tx.Date
			record.Description = tx.Description
			record.ExternalId = tx.FITID
			record.Type = 1
			record.Value = tx.Amount
			if tx.Type == "DEBIT" {
				record.Value = tx.Amount * -1
				record.Type = 2
			}

			if err := h.DB.Create(&record).Error; err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error":   "Erro ao salvar",
					"message": "Erro ao salvar a transacao " + fmt.Sprintf("%.2f", tx.Amount),
				})
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]string{
		"error":   "false",
		"message": "Arquivo OFX recebido com sucesso",
	})

	// for _, tx := range ofx.Bank.Transactions {
	// 	value, err := strconv.ParseFloat(strings.TrimSpace(tx.Amount), 64)
	// 	if err != nil {
	// 		fmt.Println("Erro ao converter valor:", err)
	// 		continue
	// 	}

	// 	date, err := time.Parse("20060102", strings.TrimSpace(tx.Date))
	// 	if err != nil {
	// 		fmt.Println("Erro ao converter data:", err)
	// 		continue
	// 	}

	// 	transaction := Transaction{
	// 		Date:        date.Format("2006-01-02"),
	// 		Type:        2,
	// 		Description: strings.TrimSpace(tx.Description),
	// 		Value:       value,
	// 		CategoryID:  1,
	// 		AccountID:   1,
	// 	}
	// 	db.Create(&transaction)
	// }
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

	// Definir ordenação padrão
	orderBy := "date DESC"

	// orderColumn := c.QueryParam("order[0][column]") // Índice da coluna
	// orderDir := c.QueryParam("order[0][dir]")       // Direção (asc ou desc)
	// if orderColumn != "" && orderDir != "" {
	// 	columns := []string{"id", "date", "amount", "category_id", "account_id"} // Defina as colunas do DataTables
	// 	colIndex, err := strconv.Atoi(orderColumn)
	// 	if err == nil && colIndex >= 0 && colIndex < len(columns) {
	// 		orderBy = columns[colIndex] + " " + orderDir
	// 	}
	// }

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

	// Prefixo do arquivo
	baseURL := "http://localhost:8000/"

	// Adicionando o prefixo ao proof
	for i := range transactions {
		if transactions[i].Proof != nil && *transactions[i].Proof != "" {
			newPath := baseURL + *transactions[i].Proof
			transactions[i].Proof = &newPath // Atualiza o ponteiro
		}
	}

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

type BalanceDTO struct {
	DateToBalance string `json:"date_to_balance" form:"date_to_balance" query:"date_to_balance"`
}

func (h *CRUDHandler) GetAccountBalance(c echo.Context) error {
	date := BalanceDTO{}

	if err := c.Bind(&date); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Dados inválidos",
			"message": err.Error(),
		})
	}

	if date.DateToBalance == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Date parameter is required",
			"message": "Informe uma data para gerar o saldo",
		})
	}

	if err := h.DB.Exec(`
		UPDATE accounts
		SET balance = (
			SELECT COALESCE(SUM(
				CASE 
					WHEN transactions.type = 1 THEN transactions.value 
					WHEN transactions.type = 2 THEN -transactions.value 
					ELSE 0 
				END
			), 0)
			FROM transactions
			WHERE transactions.account_id = accounts.id 
			AND transactions.date <= ? 
			AND transactions.deleted_at IS NULL
		)
		WHERE EXISTS (
			SELECT 1 FROM transactions 
			WHERE transactions.account_id = accounts.id
			AND transactions.date <= ?
			AND transactions.deleted_at IS NULL
		)
	`, date.DateToBalance, date.DateToBalance).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao atualizar saldo",
			"message": err.Error(),
		})
	}

	return c.NoContent(http.StatusOK)
}
