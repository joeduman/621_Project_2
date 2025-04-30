@echo off
setlocal enabledelayedexpansion

:: ------------------------------------------------
:: Prime Counter - Project 2
:: Full Auto Launcher Script
:: ------------------------------------------------

:: Kill any old running go server processes
echo [INFO] Killing previous servers...

for %%P in (fileserver.go primes_main.go primes_worker.go) do (
    for /f "tokens=2" %%a in ('tasklist /FI "IMAGENAME eq go.exe" /V /FO LIST ^| findstr /I "%%P"') do (
        taskkill /PID %%a /F >nul 2>&1
    )
)

:: ------------------------------------------------
:: Gather user experiment settings
:: ------------------------------------------------
set /p N=Enter Segment Size N (in KB): 
set /p C=Enter Chunk Size C (in KB): 
set /p M=Enter Number of Workers (M): 
set /p DATAFILE=Enter Datafile Name (e.g., numbers_1MB.dat): 
set /p CONFIGFILE=Enter Config File Name (e.g., primes_config.txt): 

set FILESERVER_LOG=fileserver.log
set MAIN_LOG=main.log

:: ------------------------------------------------
:: Clear old logs
:: ------------------------------------------------
if exist workers_combined.log del workers_combined.log
if exist worker_*.log del worker_*.log
if exist %FILESERVER_LOG% del %FILESERVER_LOG%
if exist %MAIN_LOG% del %MAIN_LOG%

:: ------------------------------------------------
:: Start Fileserver
:: ------------------------------------------------
echo [INFO] Starting Fileserver...
start "Fileserver" cmd /c "go run fileserver.go > %FILESERVER_LOG% 2>&1"
timeout /t 2 /nobreak >nul

:: ------------------------------------------------
:: Start Dispatcher + Consolidator
:: ------------------------------------------------
echo [INFO] Starting Dispatcher and Consolidator...
start "Main" cmd /c "go run primes_main.go %N% %C% %DATAFILE% %CONFIGFILE% > %MAIN_LOG% 2>&1"
timeout /t 2 /nobreak >nul

:: ------------------------------------------------
:: Start Worker Processes
:: ------------------------------------------------
echo [INFO] Launching %M% Worker Processes...
set /a count=0
set worker_pids=

:workerloop
if !count! lss %M% (
    echo [INFO] Launching Worker !count!...
    start "Worker!count!" /b cmd /c "go run primes_worker.go %C% %CONFIGFILE% > worker_!count!.log 2>&1"
    set /a count+=1
    goto workerloop
)

:: ------------------------------------------------
:: Wait for all Workers to Finish
:: ------------------------------------------------
echo [INFO] Waiting for Workers to Finish...

:waitworkers
timeout /t 5 >nul

:: Check if any "Worker" windows are still open
tasklist /v /fo table | findstr /i "Worker" >nul
if %errorlevel%==0 (
    goto waitworkers
)

:: ------------------------------------------------
:: Combine logs after all workers finish
:: ------------------------------------------------
echo [INFO] All workers finished. Combining logs...

copy /b worker_*.log workers_combined.log >nul

echo [INFO] Logs combined into workers_combined.log
echo [INFO] Completed.
exit
