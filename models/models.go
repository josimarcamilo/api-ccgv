package models

import (
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Roles admin, cc-secretaria, cc-tesouraria, cc-presidente, cc-fiscal
// é melhor criar o sistema por permissoes e deixar o usuário criar os perfis e selecionar as permissoes
type User struct {
	gorm.Model
	Name     string `gorm:"size:255;not null" json:"name" form:"name"`
	Email    string `gorm:"size:255;unique;not null" json:"email" form:"email"`
	Password string `gorm:"size:255;not null" json:"password" form:"password"`
	TeamID   uint   `gorm:"index"` // FK para Time
	Team     Team   `gorm:"foreignKey:TeamID"`
	Role     string `gorm:"size:50;null" json:"role" form:"role"`
}

type Team struct {
	ID         uint       `gorm:"primaryKey"`
	Name       string     `gorm:"size:255;not null;unique"`
	UserID     uint       `gorm:"index"`
	Users      []User     `gorm:"foreignKey:TeamID"`
	Categories []Category `gorm:"foreignKey:TeamID"`
	Accounts   []Account  `gorm:"foreignKey:TeamID"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Category struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	TeamID    uint           `gorm:"index" json:"team_id"` // FK para Time
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Account struct {
	gorm.Model
	Name    string  `gorm:"size:255;not null"`
	Balance float64 `gorm:"not null;default:0.0"`
	TeamID  uint    `gorm:"index"` // FK para Time
}

type Transaction struct {
	gorm.Model
	Type          string  `gorm:"size:50;not null"` // Ex: "entrada 1", "saída 2", "transferência 3"
	Amount        float64 `gorm:"not null"`
	Description   string  `gorm:"size:500"`
	BankAccountID uint    `gorm:"index"`                    // FK para Conta Bancária
	CategoryID    uint    `gorm:"index"`                    // FK para Categoria
	TeamID        uint    `gorm:"index"`                    // FK para Time
	Proofs        []Proof `gorm:"foreignKey:TransactionID"` // Relacionamento com comprovantes
}

// comprovantes
type Proof struct {
	gorm.Model
	TransactionID uint   `gorm:"index"`             // FK para Transação
	FilePath      string `gorm:"size:500;not null"` // Caminho do arquivo
	UploadedAt    time.Time
}

// aprovacoes tesoureiros e conselho fiscal
type Approval struct {
	gorm.Model
	TransactionID uint   `gorm:"index"` // FK para Transação
	UserID        uint   `gorm:"index"` // FK para usuário
	Status        string // aprovado, reprovado
	Observation   string
}

var storeSessions = sessions.NewCookieStore([]byte("xPjrXZsDfdlwlYzFcWZQZ92f6x9IuTkHp_m7KZTlPlg=")) // Defina uma chave secreta

// CRUD Handler Genérico
type CRUDHandler struct {
	DB        *gorm.DB
	Model     interface{}
	TableName string
}

func (h *CRUDHandler) Create(c echo.Context) error {
	// Obter a sessão
	session, err := storeSessions.Get(c.Request(), "session-id")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Sessão inválida"})
	}

	// Obter o user da sessão
	user, ok := session.Values["user"].(User)
	if !ok || user.ID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Usuário não autenticado"})
	}

	// Criar uma instância do Model dinamicamente
	record := reflect.New(reflect.TypeOf(h.Model).Elem()).Interface()

	// Fazer o bind dos dados da requisição na struct
	if err := c.Bind(record); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

	// Adicionar o TeamID do usuário autenticado
	if teamField := reflect.ValueOf(record).Elem().FieldByName("TeamID"); teamField.IsValid() && teamField.CanSet() {
		teamField.SetUint(uint64(user.TeamID)) // Define o TeamID do usuário
	}

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
