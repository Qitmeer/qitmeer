@echo off
if not exist "%~dp0qitmeer.exe" goto :error

if not exist "%~dp0config.conf" goto :config
goto :version

:config
echo ______________________________________________________________
choice /C YN /M "Qitmeer 'config.conf' file not found, Do you wan to create an new one"

IF ERRORLEVEL ==2 GOTO :no
IF ERRORLEVEL ==1 GOTO :yes
GOTO :end
:no
ECHO Fail to start Qitmeer node because the 'config.conf' file not found.
GOTO :end
:yes
ECHO Create an new Qitmeer 'config.conf' file

echo testnet=true > config.conf
echo datadir=./data >> config.conf
echo logdir=./data >> config.conf
echo listen=0.0.0.0:18130 >> config.conf
echo rpclisten=0.0.0.0:18131 >> config.conf
echo rpcuser=qitmeer >> config.conf
echo rpcpass=qitmeer123 >> config.conf
echo notls=false >> config.conf
echo printorigin=false >> config.conf
echo debuglevel=info >> config.conf

:version
echo ______________________________________________________________
echo Find Qitmeer node executable :
"%~dp0qitmeer" --version


:warnweakpass
FINDSTR "rpcuser=qitmeer" config.conf
if %errorlevel% ==0 echo WARNING using default RPC user
FINDSTR "rpcpass=qitmeer123" config.conf
if %errorlevel% ==0 echo WARNING using default RPC password


choice /C YN /M "Do you wan to start the Qitmeer node"
IF ERRORLEVEL ==2 GOTO :end
IF ERRORLEVEL ==1 GOTO :run

:run
"%~dp0qitmeer" -C=.\config.conf

echo ______________________________________________________________
echo.
goto :end
:error
echo [-] Qitmeer node executable not found.
echo Please extract all files from the downloaded package and check your package downloaded correctly from 'https://github.com/Qitmeer/qitmeer/releases'
:end
pause
