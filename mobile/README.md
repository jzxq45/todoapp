# Android APK
The APK runs the Go API and embedded SPA on a random `127.0.0.1` port. Tasks persist in `filesDir/tasks.json`; no network connection is needed.

## Requirements
- Go 1.22+, JDK 17, Android SDK Platform 35, Build Tools 35.0.0
- Android NDK 27.2.12479018 and `adb`
- Install Android Studio from https://developer.android.com/studio, or command-line tools from https://developer.android.com/tools.

## Windows SDK setup
In PowerShell, set paths after installing the SDK packages:
```powershell
$env:ANDROID_HOME = "$env:LOCALAPPDATA\Android\Sdk"
$env:ANDROID_NDK_HOME = "$env:ANDROID_HOME\ndk\27.2.12479018"
$env:JAVA_HOME = (Get-ChildItem "$env:ProgramFiles\Eclipse Adoptium\jdk-17*" -Directory | Select-Object -First 1).FullName
$env:PATH = "$env:JAVA_HOME\bin;$env:ANDROID_HOME\platform-tools;$env:PATH"
& "$env:ANDROID_HOME\cmdline-tools\latest\bin\sdkmanager.bat" "platform-tools" "platforms;android-35" "build-tools;35.0.0" "ndk;27.2.12479018"
```

## Build and install
Run from the repository root:
```sh
make deps-android
make apk
adb install -r dist/TodoApp.apk
```
APK ID: `com.example.todoapp`; minimum API: 24; ABIs: ARM, ARM64, x86, x86_64.
`make apk` creates the AAR with gomobile, rebuilds the APK, and replaces the previous `dist/TodoApp.apk`.
GitHub Actions builds the same APK for application changes and stores it as the `todoapp-apk` artifact.

## Troubleshooting
- SDK not found: set `ANDROID_HOME` to the SDK directory.
- NDK not found: set `ANDROID_NDK_HOME` to NDK `27.2.12479018`.
- Gradle Java version error: select JDK 17 in `JAVA_HOME`.
- Server logs: `adb shell run-as com.example.todoapp cat files/app.log`.
