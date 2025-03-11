package main

import (
	"encoding/gob"
	"html/template"
	"io"
	"jc-financas/models"
	"jc-financas/repositories"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	// "github.com/labstack/echo/middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// Criar um novo store de sessões
var storeSessions = sessions.NewCookieStore([]byte("xPjrXZsDfdlwlYzFcWZQZ92f6x9IuTkHp_m7KZTlPlg=")) // Defina uma chave secreta

func main() {
	gob.Register(models.User{})

	storeSessions.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600, // O cookie vai durar 1 hora
		HttpOnly: true, // Impede o acesso ao cookie via JavaScript
	}
	// Inicializar o banco de dados
	db, err := gorm.Open(sqlite.Open("financas.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}

	// Migrar o modelo para o banco de dados
	if err := db.AutoMigrate(
		&models.User{},
		&models.Account{},
		&models.Team{},
		&models.Category{},
		&models.Transaction{},
	); err != nil {
		log.Fatalf("Erro ao migrar o banco de dados: %v", err)
	}

	// Inicializar o Echo
	e := echo.New()
	t := &Template{
		templates: template.Must(template.ParseGlob("./*.html")),
	}
	e.Renderer = t

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://localhost:8000",
			"https://app.orfed.com.br",
		},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true, // Necessário para cookies
	}))

	// Lida manualmente com OPTIONS caso necessário
	e.OPTIONS("/*", func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "https://app.orfed.com.br")
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
		return c.NoContent(204)
	})

	e.Static("/static", "static")

	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/home")
	})
	e.GET("/login", Login)
	e.GET("/register", Register)

	userHandler := repositories.CRUDHandler{DB: db, Model: &models.User{}, TableName: "users"}

	e.POST("/register", userHandler.Register)

	e.GET("/home", Home)

	e.POST("/login", func(c echo.Context) error {
		type LoginRequest struct {
			Email    string `json:"email" form:"email"`
			Password string `json:"password" form:"password"`
		}

		var loginRequest LoginRequest
		if err := c.Bind(&loginRequest); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
		}

		// Verificar se o e-mail e senha foram fornecidos
		if loginRequest.Email == "" || loginRequest.Password == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email e senha são obrigatórios"})
		}

		// Buscar o usuário pelo e-mail
		var user models.User
		if err := db.Where("email = ?", loginRequest.Email).First(&user).Error; err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Credenciais inválidas"})
		}

		// Comparar a senha fornecida com o hash armazenado
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Credenciais inválidas"})
		}

		// Criar uma nova sessão
		session, _ := storeSessions.Get(c.Request(), "session-id")
		session.Values["user"] = user
		session.Save(c.Request(), c.Response())

		err = session.Save(c.Request(), c.Response())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"erro ao salvar sessao": err.Error()})
		}

		return c.Redirect(http.StatusSeeOther, "/home")
	})

	e.GET("/profile", func(c echo.Context) error {
		session, err := storeSessions.Get(c.Request(), "session-id")

		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
		}
		userLogged, ok := session.Values["user"]

		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Nao autenticado"})
		}

		return c.JSON(http.StatusOK, userLogged)
	})

	e.GET("/logout", func(c echo.Context) error {
		session, _ := storeSessions.Get(c.Request(), "session-id")
		// Remover a chave da sessão
		session.Values["userID"] = nil
		session.Save(c.Request(), c.Response())

		return c.JSON(http.StatusOK, map[string]string{"message": "Logout realizado com sucesso"})
	})

	e.GET("/teams/create", TeamCreate)
	e.GET("/teams", func(c echo.Context) error {
		var model []models.Team
		// db.Preload("Users").Find(&model).Error;
		if err := db.Find(&model).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar times"})
		}
		return c.JSON(http.StatusOK, model)
	})
	e.POST("/teams", func(c echo.Context) error {
		// Obter a sessão
		session, err := storeSessions.Get(c.Request(), "session-id")
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Sessão inválida"})
		}

		// Obter o userID da sessão
		user, ok := session.Values["user"].(models.User)
		if !ok || user.ID == 0 {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Usuário não autenticado"})
		}

		// Bind do nome do time
		var requestBody struct {
			Name string `json:"name" form:"name"`
		}
		if err := c.Bind(&requestBody); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
		}

		if requestBody.Name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "O nome do time é obrigatório"})
		}

		// Criar o time
		team := models.Team{
			Name:   requestBody.Name,
			UserID: user.ID,
		}
		if err := db.Create(&team).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao criar o time"})
		}

		// Atualizar o team_id do usuário logado
		if err := db.Model(&models.User{}).Where("id = ?", user.ID).Update("team_id", team.ID).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao associar o time ao usuário"})
		}

		user.TeamID = team.ID
		session.Values["user"] = user
		session.Save(c.Request(), c.Response())

		err = session.Save(c.Request(), c.Response())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error atualizar sessao": err.Error()})
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message": "Time criado com sucesso",
			"team":    team,
		})
	})

	e.GET("/crud-config/:entity", GetCrudConfig)
	e.GET("/crud/:entity", Crud)

	categoryHandler := repositories.CRUDHandler{DB: db, Model: &models.Category{}, TableName: "categories"}
	accountHandler := repositories.CRUDHandler{DB: db, Model: &models.Account{}, TableName: "accounts"}
	transactionHandler := repositories.CRUDHandler{DB: db, Model: &models.Transaction{}, TableName: "transactions"}

	// Registrar rotas
	e.POST("/categories", categoryHandler.CreateCategory)
	e.GET("/categories", categoryHandler.ListCategories)
	e.GET("/categories/edit/:id", FormCategoryEdit)
	e.GET("/categories/:id", categoryHandler.GetByID)
	e.PUT("/categories/:id", categoryHandler.Update)
	e.DELETE("/categories/:id", categoryHandler.Delete)

	e.GET("/users/table", ListUsers)
	e.POST("/users", userHandler.CreateUser)
	e.GET("/users", userHandler.ListUsers)
	// e.GET("/users/:id", userHandler.GetByID)
	// e.PUT("/users/:id", userHandler.Update)
	// e.DELETE("/users/:id", userHandler.Delete)

	e.GET("/categories/table", ListCategories)

	// contas
	e.GET("/accounts/table", ListAccounts)
	e.POST("/accounts", accountHandler.CreateAccount)
	e.GET("/accounts", accountHandler.ListAccounts)
	e.GET("/accounts/:id", accountHandler.GetByID)
	e.PUT("/accounts/:id", accountHandler.UpdateAccount)
	e.GET("/accounts/edit/:id", FormAccountEdit)
	e.DELETE("/accounts/:id", accountHandler.Delete)
	e.POST("/account-balance", accountHandler.GetAccountBalance)

	e.GET("/transactions/create", CreateTransaction)
	e.GET("/transactions/table", ListTransactions)
	e.POST("/transactions", transactionHandler.CreateTransaction)
	e.GET("/transactions", transactionHandler.ListTransactions)
	e.GET("/transactions/:id", transactionHandler.GetByID)
	e.PUT("/transactions/:id", transactionHandler.Update)
	e.DELETE("/transactions/:id", transactionHandler.Delete)
	e.GET("/transactions/import", ImportTransaction)
	e.POST("/transactions/import-ofx", transactionHandler.ImportOFX)
	e.POST("/transactions/import-csv", transactionHandler.ImportCSV)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // Porta padrão para desenvolvimento local
	}
	e.Logger.Fatal(e.Start(":" + port))
}

func GetCrudConfig(c echo.Context) error {
	entity := c.Param("entity")

	configs := map[string]interface{}{
		"users": map[string]interface{}{
			"entity": "users",
			"title":  "Usuários",
			"apiUrl": "http://localhost:8000/users",
			"fields": []map[string]interface{}{
				{"name": "name", "label": "Nome", "data": "name", "type": "text", "required": true},
				{"name": "email", "label": "E-mail", "data": "email", "type": "mail", "required": true},
			},
		},
		"categories": map[string]interface{}{
			"entity":   "categories",
			"title":    "Categorias",
			"apiUrl":   "http://localhost:8000/categories",
			"formEdit": "http://localhost:8000/categories/edit/",
			"fields": []map[string]interface{}{
				{"name": "id", "label": "ID", "data": "id", "type": "number", "readonly": true},
				{"name": "name", "label": "Nome", "data": "name", "type": "text", "required": true},
			},
		},
		"transactionscreate": map[string]interface{}{
			"apiUrlCategories":  "http://localhost:8000/categories?toselect=true",
			"apiUrlAccounts":    "http://localhost:8000/accounts?toselect=true",
			"apiUrlTransaction": "http://localhost:8000/transactions",
			"apiUrlImportOfx":   "http://localhost:8000/transactions/import-ofx",
			"apiUrlImportCsv":   "http://localhost:8000/transactions/import-csv",
		},
		"accounts": map[string]interface{}{
			"entity":        "accounts",
			"title":         "Contas Contábeis",
			"apiUrl":        "http://localhost:8000/accounts",
			"urlGetBalance": "http://localhost:8000/account-balance",
			"formEdit":      "http://localhost:8000/accounts/edit/",
			"fields": []map[string]interface{}{
				{"name": "id", "label": "ID", "data": "id", "type": "number", "readonly": true},
				{"name": "name", "label": "Nome", "data": "name", "type": "text", "required": true},
				{"name": "balance", "label": "Saldo", "data": "balance", "type": "text", "readonly": true},
			},
		},
		"transactions": map[string]interface{}{
			"entity": "transactions",
			"title":  "Transações",
			"apiUrl": "http://localhost:8000/transactions",
			"transaction_types": []map[string]interface{}{
				{"value": 1, "label": "Entrada"},
				{"value": 2, "label": "Saída"},
				{"value": 3, "label": "Transferência"},
			},
			"fields": []map[string]interface{}{
				{"name": "id", "label": "ID", "data": "id", "type": "number", "readonly": true},
				{"name": "date_at", "label": "Data", "data": "date_at", "type": "date", "required": true},
				{"name": "type", "label": "Tipo", "data": "type", "type": "select", "options": []map[string]interface{}{
					{"value": 1, "label": "Entrada"},
					{"value": 2, "label": "Saída"},
				}, "required": true},
				{"name": "description", "label": "Descrição", "data": "description", "type": "text", "required": true},
				{"name": "value", "label": "Valor", "data": "value", "type": "text", "required": true},
				{"name": "category_id", "label": "Categoria", "data": "category.name", "type": "select", "source": "/categories", "required": true},
				{"name": "account_id", "label": "Conta", "data": "account.name", "type": "select", "source": "/accounts", "required": true},
				{"name": "proof", "label": "Comprovante", "data": "proof", "type": "file"},
			},
		},
	}

	if config, exists := configs[entity]; exists {
		return c.JSON(http.StatusOK, config)
	}
	return c.JSON(http.StatusNotFound, map[string]string{"error": "Configuração não encontrada"})
}

func Crud(c echo.Context) error {
	entity := c.Param("entity")
	return c.Render(http.StatusOK, "crud", map[string]interface{}{
		"Entity":       entity,
		"CurrentRoute": "/crud/" + entity,
	})
}

func ListTransactions(c echo.Context) error {
	err := VerifySession(c)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "transactions", map[string]interface{}{
		"Entity":       "transactions",
		"CurrentRoute": "/transactions/table",
	})
}

func CreateTransaction(c echo.Context) error {
	err := VerifySession(c)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "transactions-create", map[string]interface{}{
		"Entity":       "transactions",
		"CurrentRoute": "/transactions/create",
	})
}

func ImportTransaction(c echo.Context) error {
	err := VerifySession(c)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "transactions-import", map[string]interface{}{
		"Entity":       "transactions",
		"CurrentRoute": "/transactions/import",
	})
}

func ListUsers(c echo.Context) error {
	err := VerifySession(c)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "users", map[string]interface{}{
		"Entity":       "users",
		"CurrentRoute": "/users/table",
	})
}

func ListCategories(c echo.Context) error {
	err := VerifySession(c)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "categories", map[string]interface{}{
		"Entity":       "categories",
		"CurrentRoute": "/categories/table",
	})
}

func FormCategoryEdit(c echo.Context) error {
	err := VerifySession(c)
	if err != nil {
		return err
	}
	id := c.Param("id")

	return c.Render(http.StatusOK, "categories-edit", map[string]interface{}{
		"Entity":       "categories",
		"EntityId":     id,
		"CurrentRoute": "/categories/edit",
	})
}

func FormAccountEdit(c echo.Context) error {
	err := VerifySession(c)
	if err != nil {
		return err
	}
	id := c.Param("id")

	return c.Render(http.StatusOK, "accounts-edit", map[string]interface{}{
		"Entity":       "accounts",
		"EntityId":     id,
		"CurrentRoute": "/accounts/edit",
	})
}

func ListAccounts(c echo.Context) error {
	err := VerifySession(c)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "accounts", map[string]interface{}{
		"Entity":       "accounts",
		"CurrentRoute": "/accounts/table",
	})
}

func Hello(c echo.Context) error {
	return c.Render(http.StatusOK, "hello", "World")
}

func Login(c echo.Context) error {
	return c.Render(http.StatusOK, "login", "World")
}

func Home(c echo.Context) error {
	err := VerifySession(c)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "home", map[string]interface{}{
		"CurrentRoute": "/home",
	})
}

func Register(c echo.Context) error {
	return c.Render(http.StatusOK, "register", "World")
}

func AccountCreate(c echo.Context) error {
	return c.Render(http.StatusOK, "account", "World")
}

func TeamCreate(c echo.Context) error {

	return c.Render(http.StatusOK, "team-create", map[string]interface{}{
		"CurrentRoute": "/teams/create",
	})
}

func AccountList(c echo.Context) error {
	session, err := storeSessions.Get(c.Request(), "session-id")
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	// Obter o user_id da sessão
	userID, ok := session.Values["userID"].(uint)
	if !ok || userID == 0 {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	return c.Render(http.StatusOK, "account-list", "World")
}

func VerifySession(c echo.Context) error {
	session, err := storeSessions.Get(c.Request(), "session-id")
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	user, ok := session.Values["user"].(models.User)
	if !ok || user.ID == 0 {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	return nil
}
