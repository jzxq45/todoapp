# TodoApp

Список задач на Go с REST API и встроенным веб-интерфейсом.

```sh
make run
```

Откройте http://localhost:8080. По умолчанию данные хранятся в памяти.

```sh
make test
```

API: `/healthz`, `/api/tasks`.

Запуск в Docker: `docker compose up --build`; порт публикуется только на localhost.

APK: `make apk` (Android SDK/JDK 17); файл `dist/TodoApp.apk`.
GitHub Actions также собирает APK при изменениях приложения.
