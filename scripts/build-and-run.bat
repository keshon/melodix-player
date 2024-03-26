@echo off

rem
rem BUILD
rem

rem Get Go version
for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i

rem Get the build date
for /f "tokens=*" %%a in ('powershell -command "Get-Date -UFormat '%%Y-%%m-%%dT%%H:%%M:%%SZ'"') do set BUILD_DATE=%%a

rem Navigate to the root of the project from the scripts folder
cd ..

rem Build command
go run -ldflags "-X github.com/keshon/melodix-player/internal/version.BuildDate=%BUILD_DATE% -X github.com/keshon/melodix-player/internal/version.GoVersion=%GO_VERSION%" cmd\melodix\melodix.go

rem Return to the scripts folder after execution
cd scripts