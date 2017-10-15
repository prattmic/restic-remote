Dim WinScriptHost
Set WinScriptHost = CreateObject("WScript.Shell")
' Run the client and hide the window (final argument)
WinScriptHost.Run Chr(34) & "%APPDATA%\restic-remote\client.exe" & Chr(34), 0
Set WinScriptHost = Nothing
