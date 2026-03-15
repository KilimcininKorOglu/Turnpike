@echo off
REM Turnpike - Build Script for Windows
REM Usage: build.bat [command]

setlocal enabledelayedexpansion

set APP_NAME=turnpike
set VERSION=2.0.2
set BUILD_DIR=build
set CMD_PATH=.\cmd\turnpike
set LDFLAGS=-s -w -X "github.com/KilimcininKorOglu/Turnpike/internal/cli.AppVersion=%VERSION%"

if "%~1"=="" goto all
if "%~1"=="all" goto all
if "%~1"=="build" goto build
if "%~1"=="build-all" goto build_all
if "%~1"=="build-windows" goto build_windows
if "%~1"=="build-darwin" goto build_darwin
if "%~1"=="build-linux" goto build_linux
if "%~1"=="test" goto test
if "%~1"=="test-verbose" goto test_verbose
if "%~1"=="test-race" goto test_race
if "%~1"=="test-cover" goto test_cover
if "%~1"=="vet" goto vet
if "%~1"=="fmt" goto fmt
if "%~1"=="tidy" goto tidy
if "%~1"=="lint" goto lint
if "%~1"=="run" goto run
if "%~1"=="run-cli" goto run_cli
if "%~1"=="clean" goto clean
if "%~1"=="version" goto version
if "%~1"=="help" goto help
echo Unknown command: %~1
goto help

REM ─────────────────────────────────────────────────
REM Default: clean + build all platforms
REM ─────────────────────────────────────────────────

:all
call :clean
call :build_all
goto end

REM ─────────────────────────────────────────────────
REM Build targets
REM ─────────────────────────────────────────────────

:build
for /f "tokens=*" %%a in ('go env GOARCH') do set CURRENT_ARCH=%%a
echo Building %APP_NAME% for windows/%CURRENT_ARCH%...
if not exist %BUILD_DIR% mkdir %BUILD_DIR%
go build -ldflags "%LDFLAGS%" -o %BUILD_DIR%\%APP_NAME%-windows-%CURRENT_ARCH%.exe %CMD_PATH%
if %ERRORLEVEL% equ 0 (
    echo Output: %BUILD_DIR%\%APP_NAME%-windows-%CURRENT_ARCH%.exe
) else (
    echo BUILD FAILED
    exit /b 1
)
goto end

:build_all
if not exist %BUILD_DIR% mkdir %BUILD_DIR%

call :build_platform windows amd64 .exe
call :build_platform windows arm64 .exe
call :build_platform darwin amd64 ""
call :build_platform darwin arm64 ""
call :build_platform linux amd64 ""
call :build_platform linux arm64 ""

echo.
echo All builds complete:
dir /b %BUILD_DIR%\%APP_NAME%-* 2>nul
goto end

:build_windows
if not exist %BUILD_DIR% mkdir %BUILD_DIR%
call :build_platform windows amd64 .exe
call :build_platform windows arm64 .exe
goto end

:build_darwin
if not exist %BUILD_DIR% mkdir %BUILD_DIR%
call :build_platform darwin amd64 ""
call :build_platform darwin arm64 ""
goto end

:build_linux
if not exist %BUILD_DIR% mkdir %BUILD_DIR%
call :build_platform linux amd64 ""
call :build_platform linux arm64 ""
goto end

:build_platform
set GOOS=%~1
set GOARCH=%~2
set EXT=%~3
set OUTPUT=%BUILD_DIR%\%APP_NAME%-%GOOS%-%GOARCH%%EXT%
echo Building %OUTPUT%...
set CGO_ENABLED=0
go build -ldflags "%LDFLAGS%" -o %OUTPUT% %CMD_PATH% 2>nul
if %ERRORLEVEL% neq 0 (
    echo   Skipped %GOOS%/%GOARCH% ^(cross-compilation not available^)
)
goto :eof

REM ─────────────────────────────────────────────────
REM Test targets
REM ─────────────────────────────────────────────────

:test
echo Running tests...
go test ./internal/... -count=1
goto end

:test_verbose
go test ./internal/... -count=1 -v
goto end

:test_race
go test ./internal/... -count=1 -race
goto end

:test_cover
if not exist %BUILD_DIR% mkdir %BUILD_DIR%
go test ./internal/... -count=1 -coverprofile=%BUILD_DIR%\coverage.out
go tool cover -func=%BUILD_DIR%\coverage.out
echo.
echo HTML report: %BUILD_DIR%\coverage.html
go tool cover -html=%BUILD_DIR%\coverage.out -o %BUILD_DIR%\coverage.html
goto end

REM ─────────────────────────────────────────────────
REM Code quality
REM ─────────────────────────────────────────────────

:vet
go vet ./...
goto end

:fmt
go fmt ./...
goto end

:tidy
go mod tidy
goto end

:lint
call :vet
call :fmt
goto end

REM ─────────────────────────────────────────────────
REM Run & utility
REM ─────────────────────────────────────────────────

:run
go run -ldflags "%LDFLAGS%" %CMD_PATH%
goto end

:run_cli
go run -ldflags "%LDFLAGS%" %CMD_PATH% --version
goto end

:clean
echo Cleaning build directory...
if exist %BUILD_DIR% rmdir /s /q %BUILD_DIR%
echo Clean complete.
goto end

:version
echo %APP_NAME% v%VERSION%
goto end

REM ─────────────────────────────────────────────────
REM Help
REM ─────────────────────────────────────────────────

:help
echo %APP_NAME% v%VERSION% - Build Script
echo.
echo Usage: build.bat [command]
echo.
echo Build:
echo   build          Build for current platform
echo   build-all      Build for all platforms (default)
echo   build-windows  Build for Windows (amd64 + arm64)
echo   build-darwin   Build for macOS (amd64 + arm64)
echo   build-linux    Build for Linux (amd64 + arm64)
echo.
echo Test:
echo   test           Run all tests
echo   test-verbose   Run tests with verbose output
echo   test-race      Run tests with race detector
echo   test-cover     Run tests with coverage report
echo.
echo Quality:
echo   vet            Run go vet
echo   fmt            Format code
echo   tidy           Tidy module dependencies
echo   lint           Run all linters (vet + fmt)
echo.
echo Utility:
echo   run            Run the application (GUI mode)
echo   run-cli        Run CLI version check
echo   clean          Remove build artifacts
echo   version        Show version
echo   help           Show this help
goto end

:end
endlocal
