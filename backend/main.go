package main

func main() {
	// 1. env configuration
	env := EnvConfig()
	// 2. database connection
	dbConn := DBConnection(env)
}
