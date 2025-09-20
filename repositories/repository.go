package repositories

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"jc-financas/consts"
	"jc-financas/helpers"
	"jc-financas/models"
	"jc-financas/services/ofx"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var storeSessions = sessions.NewCookieStore([]byte("xPjrXZsDfdlwlYzFcWZQZ92f6x9IuTkHp_m7KZTlPlg=")) // Defina uma chave secreta
var DB *gorm.DB

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
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name, email e senha são obrigatórios"})
	}

	// Encriptar a senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao encriptar senha"})
	}

	// Substituir a senha em texto puro pela senha encriptada
	user.Password = string(hashedPassword)
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

	if defaultTeam.ID == 0 {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Time nao criado",
			"message": "",
		})
	}

	// Atualizar o team_id do usuário logado
	if err := h.DB.Model(&models.User{}).Where("id = ?", user.ID).Update("team_id", defaultTeam.ID).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao associar o time ao usuário"})
	}

	user.TeamID = &defaultTeam.ID
	return c.JSON(http.StatusCreated, user)
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
	claims, err := ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	model := models.Category{}
	if err := c.Bind(&model); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

	// Validações simples
	if model.Name == "" || model.Type == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "name and type are required",
		})
	}

	// Definir TeamID
	model.TeamID = claims.TeamID

	// Salvar o registro no banco
	if err := h.DB.Create(&model).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar categoria"})
	}

	return c.JSON(http.StatusCreated, model)
}

func (h *CRUDHandler) ListCategories(c echo.Context) error {
	claims, err := ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	query := h.DB.
		Where("team_id = ?", claims.TeamID)

	typeParam := c.QueryParam("type")
	useMapParam := c.QueryParam("use_map")
	if typeParam != "" {
		query.Where("type = ?", typeParam)
	}
	if useMapParam != "" {
		query.Where("use_map = ?", useMapParam)
	}

	var records []models.Category
	if err := query.Order("id ASC").Find(&records).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao buscar registros para select",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, records)
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

	if true {
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
	model.TeamID = *user.TeamID

	// Salvar o registro no banco
	if err := h.DB.Create(&model).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar"})
	}

	return c.JSON(http.StatusCreated, model)
}

// func (h *CRUDHandler) CreateTransaction(c echo.Context) error {
// 	// Obter a sessão e o usuário logado
// 	session, err := storeSessions.Get(c.Request(), "session-id")
// 	if err != nil {
// 		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Sessão inválida"})
// 	}
// 	user, ok := session.Values["user"].(models.User)
// 	if !ok || user.ID == 0 {
// 		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Usuário não autenticado"})
// 	}

// 	newTransaction := models.Transaction{}
// 	newTransaction.TeamID = user.TeamID

// 	typeParam := c.FormValue("date")
// 	if typeParam == "1" {
// 		// newTransaction.Type, _ = strconv.Atoi(typeParam)
// 	}

// 	if err := c.Bind(&newTransaction); err != nil {
// 		return c.JSON(http.StatusBadRequest, map[string]string{
// 			"error":   "Dados inválidos",
// 			"message": err.Error(),
// 		})
// 	}

// 	file, err := c.FormFile("proof")
// 	if err == nil {
// 		src, err := file.Open()
// 		if err != nil {
// 			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao abrir o arquivo"})
// 		}
// 		defer src.Close()
// 		dateStr := c.FormValue("date")

// 		t, err := time.Parse("2006-01-02", dateStr)
// 		if err != nil {
// 			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro converter a data"})
// 		}

// 		// Criar o diretório do ano/mês
// 		anoMes := t.Format("2006/01")
// 		dir := fmt.Sprintf("static/comprovantes/%v/", anoMes)
// 		os.MkdirAll(dir, os.ModePerm)

// 		// Gerar o caminho do arquivo
// 		filename := fmt.Sprintf("%s%s", uuid.New().String(), filepath.Ext(file.Filename))
// 		dstPath := filepath.Join(dir, filename)

// 		// Criar o arquivo no servidor
// 		dst, err := os.Create(dstPath)
// 		if err != nil {
// 			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar o arquivo"})
// 		}
// 		defer dst.Close()
// 		io.Copy(dst, src)

// 		// Salvar o caminho do comprovante
// 		newTransaction.ReceiptUrl = &dstPath
// 	}

// 	// Salvar a transação no banco
// 	if err := h.DB.Create(&newTransaction).Error; err != nil {
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar"})
// 	}

// 	accountDestination := c.FormValue("account_destination")

// 	if accountDestination != "" {
// 		acDes, _ := strconv.ParseUint(accountDestination, 10, 64)
// 		var accDestination uint = uint(acDes)

// 		transactionDestination := models.Transaction{}
// 		transactionDestination.TeamID = newTransaction.TeamID
// 		transactionDestination.Date = newTransaction.Date
// 		transactionDestination.Type = models.TransactionTypeEntrada
// 		transactionDestination.Description = newTransaction.Description
// 		transactionDestination.CategoryID = newTransaction.CategoryID
// 		transactionDestination.AccountID = &accDestination
// 		transactionDestination.TransactionOriginId = &newTransaction.ID
// 		transactionDestination.Value = newTransaction.Value

// 		if err := h.DB.Create(&transactionDestination).Error; err != nil {
// 			return c.JSON(http.StatusInternalServerError, map[string]string{
// 				"error":   "Erro ao salvar",
// 				"message": "Erro ao salvar na conta de destino",
// 				"detail":  err.Error(),
// 			})
// 		}

// 	}

// 	return c.JSON(http.StatusCreated, newTransaction)
// }

type TransactionOFX struct {
	Type        string
	Date        string
	Amount      string
	FITID       string
	CheckNum    string
	Description string
}

type AccountImportOFX struct {
	AccountID uint                  `json:"account_id" form:"account_id"`
	Bank      string                `json:"bank" form:"bank"` // bradesco, inter
	File      *multipart.FileHeader `json:"file" form:"file"`
}

func (h *CRUDHandler) ImportOFX(c echo.Context) error {
	claims, err := ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	var dto AccountImportOFX
	if err := c.Bind(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Erro ao interpretar a requisição",
			"message": err.Error(),
		})
	}

	if dto.Bank == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Banco obrigatório",
		})
	}

	if dto.AccountID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Conta obrigatória",
		})
	}

	account, err := GetAccount(dto.AccountID, claims.TeamID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Conta não encontrada",
			"message": err.Error(),
		})
	}
	if account.TeamID != claims.TeamID {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Conta não pertence à equipe",
		})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro",
			"message": "Erro ao ler o arquivo",
		})
	}
	if dto.Bank == "inter" {
		transactions, err := ofx.TranslaterOfxInter(file, *account)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error":   dto.Bank + " Erro ao processar arquivo OFX",
				"message": err.Error(),
			})
		}

		for _, tx := range transactions {
			exists := models.Transaction{}
			if err := h.DB.Where("team_id = ?", claims.TeamID).Where("external_id = ?", tx.ExternalId).First(&exists).Error; err != nil {
				if err := h.DB.Create(&tx).Error; err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{
						"error":   "Erro ao salvar",
						"message": "Erro ao salvar a transacao " + fmt.Sprintf("%v", tx.Value),
					})
				}
			}
		}

		return c.JSON(http.StatusOK, map[string]string{
			"error":   "false",
			"message": dto.Bank + " Arquivo OFX recebido com sucesso",
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
			current.Amount = strings.TrimPrefix(line, "<TRNAMT>")
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

	fmt.Printf("%v %v %v %v %v\n", "Tipo", "Data", "Valor", "Código", "Descrição")
	fmt.Println(strings.Repeat("-", 70))

	for _, tx := range transactions {
		fmt.Printf("%v %v %v %v %v\n", tx.Type, tx.Date, tx.Amount, tx.FITID, tx.Description)
		record := models.Transaction{}
		if err := h.DB.Where("team_id = ?", claims.TeamID).Where("external_id = ?", tx.FITID).First(&record).Error; err != nil {
			//criar
			record.TeamID = claims.TeamID
			record.Date = tx.Date
			record.Description = tx.Description
			record.ExternalId = &tx.FITID
			record.Type = 1
			record.AccountID = &dto.AccountID
			record.Value, err = helpers.OnlyNumbers(tx.Amount)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error":   "Erro OnlyNumbers",
					"message": err.Error(),
				})
			}

			if tx.Type == "DEBIT" {
				record.Type = 2
			}

			if err := h.DB.Create(&record).Error; err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error":   "Erro ao salvar",
					"message": "Erro ao salvar a transacao " + fmt.Sprintf("%v", tx.Amount),
				})
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]string{
		"error":   "false",
		"message": "Arquivo OFX recebido com sucesso",
	})
}

type AccountImportCSV struct {
	AccountID uint                  `json:"account_id" form:"account_id"`
	File      *multipart.FileHeader `json:"file" form:"file"`
}

func (h *CRUDHandler) ImportCSV(c echo.Context) error {
	claims, err := ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	var dto AccountImportCSV
	if err := c.Bind(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Erro ao interpretar a requisição",
			"message": err.Error(),
		})
	}

	account, err := GetAccount(dto.AccountID, claims.TeamID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Conta não encontrada",
			"message": err.Error(),
		})
	}

	if account.TeamID != claims.TeamID {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Conta não pertence à equipe",
		})
	}

	src, err := dto.File.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao abrir o arquivo",
			"message": err.Error(),
		})
	}
	defer src.Close()

	// Criar um arquivo temporário para leitura
	tempFile, err := os.CreateTemp("", "upload-*.csv")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao criar arquivo temporário",
			"message": err.Error(),
		})
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.ReadFrom(src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao salvar arquivo temporário",
			"message": err.Error(),
		})
	}
	tempFile.Close()

	// Abrir e ler o arquivo CSV
	csvFile, err := os.Open(tempFile.Name())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao ler o arquivo CSV",
			"message": err.Error(),
		})
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	reader.FieldsPerRecord = -1 // Permite linhas com números diferentes de colunas
	// reader.Comma = ';'          // Define o delimitador ; padrao é ,

	rows, err := reader.ReadAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao processar o arquivo CSV",
			"message": err.Error(),
		})
	}

	// Criar uma tabela formatada no console
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"DATA", "CATEGORIA", "CAT. MAPA", "HISTÓRICO", "ENTRADA", "SAÍDA"})

	const COLUMN_DATE = 0
	const COLUMN_CAT = 1
	const COLUMN_CAT_MAP = 2
	const COLUMN_HISTORY = 3
	const COLUMN_ENTRY = 4
	const COLUMN_EXIT = 5
	for i, row := range rows {
		// verifique se row[COLUMN_DATE] contem uma data válida
		var dateRow time.Time
		dateRow, err := time.Parse("02/01/2006", row[COLUMN_DATE])

		if err != nil {
			fmt.Println("Data inválida na linha", i+1, ":", row[COLUMN_DATE])
			continue
		}

		if len(row) < 6 {
			continue // Garantir que há colunas suficientes
		}

		table.Append(row[:6])

		if row[COLUMN_ENTRY] == "" && row[COLUMN_EXIT] == "" {
			continue
		}
		category := models.Category{TeamID: claims.TeamID, Name: row[COLUMN_CAT], UseMap: false}

		record := models.Transaction{}
		record.TeamID = claims.TeamID
		record.Date = dateRow.Format("2006-01-02")
		record.Description = row[COLUMN_HISTORY]
		record.AccountID = &dto.AccountID

		if row[COLUMN_ENTRY] != "" {
			record.Type = consts.TransactionTypeEntrada
			category.Type = consts.CategoryTypeEntry
			record.Value, err = helpers.OnlyNumbers(row[COLUMN_ENTRY])
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error":   "Erro OnlyNumbers",
					"message": err.Error(),
				})
			}

		}

		if row[COLUMN_EXIT] != "" {
			record.Type = consts.TransactionTypeSaida
			category.Type = consts.CategoryTypeExit
			record.Value, err = helpers.OnlyNumbers(row[COLUMN_EXIT])
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error":   "Erro OnlyNumbers",
					"message": err.Error(),
				})
			}
		}

		if row[COLUMN_CAT_MAP] != "" {
			categoryMap := models.Category{Type: record.Type, TeamID: claims.TeamID, Name: row[COLUMN_CAT_MAP], UseMap: true}
			newCatMap, err := UpsertCategory(&categoryMap)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error":   "Erro ao salvar a categoria do mapa.",
					"message": err.Error(),
				})
			}
			record.CategoryMapID = &newCatMap.ID
		}

		newCat, err := UpsertCategory(&category)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error":   "Erro ao salvar a categoria.",
				"message": err.Error(),
			})
		}

		fmt.Printf("categoria cadastrada %d", newCat.ID)
		record.CategoryID = &newCat.ID

		if err := h.DB.Create(&record).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error":   "Erro ao salvar a transacao",
				"message": err.Error(),
			})
		}
	}
	table.Render()

	return c.JSON(http.StatusOK, map[string]string{
		"error":   "false",
		"message": "Arquivo CSV recebido com sucesso",
	})
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
	// start, _ := strconv.Atoi(c.QueryParam("start"))
	// length, _ := strconv.Atoi(c.QueryParam("length"))
	// search := c.QueryParam("search[value]")

	// // Definir ordenação padrão
	// orderBy := "date DESC"

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
	// if search != "" {
	// 	query = query.Where("description LIKE ? OR amount::text LIKE ?", "%"+search+"%", "%"+search+"%")
	// }

	// Aplicar paginação e ordenação
	var transactions []models.Transaction
	if err := query.Limit(100).Find(&transactions).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar registros"})
	}

	return c.JSON(http.StatusOK, transactions)

	// Contar total de registros sem filtros
	// var totalRecords int64
	// h.DB.Model(&models.Transaction{}).Where("team_id = ?", user.TeamID).Count(&totalRecords)

	// Contar total de registros filtrados
	// totalFiltered := totalRecords

	// Prefixo do arquivo
	// baseURL := "http://localhost:8000/"

	// // Adicionando o prefixo ao proof
	// for i := range transactions {
	// 	if transactions[i].Proof != nil && *transactions[i].Proof != "" {
	// 		newPath := baseURL + *transactions[i].Proof
	// 		transactions[i].Proof = &newPath // Atualiza o ponteiro
	// 	}
	// }

	// // Retornar resposta no formato esperado pelo DataTables
	// response := map[string]interface{}{
	// 	"draw":            c.QueryParam("draw"),
	// 	"recordsTotal":    totalRecords,
	// 	"recordsFiltered": totalFiltered,
	// 	"data":            transactions,
	// }

	// return c.JSON(http.StatusOK, response)
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
