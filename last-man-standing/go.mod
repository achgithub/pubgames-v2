module pubgames/last-man-standing

go 1.25

require (
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/mux v1.8.1
	github.com/mattn/go-sqlite3 v1.14.33
	golang.org/x/crypto v0.46.0
	pubgames/shared/auth v0.0.0
	pubgames/shared/config v0.0.0
)

require github.com/felixge/httpsnoop v1.0.3 // indirect

replace pubgames/shared/auth => ../shared/auth

replace pubgames/shared/config => ../shared/config
