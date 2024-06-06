module github.com/nesiler/cestx/test

go 1.22

require (
	github.com/joho/godotenv v1.5.1
	github.com/nesiler/cestx/common v0.0.0
)

require (
	github.com/fatih/color v1.17.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.18.0 // indirect
)

// for local development
replace github.com/nesiler/cestx/common => ../common
