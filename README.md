[![Go Report Card](https://goreportcard.com/badge/github.com/duke0x/ts-notifier)](https://goreportcard.com/report/github.com/duke0x/ts-notifier) [![Go Reference](https://pkg.go.dev/badge/github.com/duke0x/ts-notifier.svg)](https://pkg.go.dev/github.com/duke0x/ts-notifier)
![Coverage](https://img.shields.io/badge/Coverage-79.7%25-yellow)

# Time Spend Notifier

Эта утилита предназначена для подсчета отработанного времени каждого участника команды и 
информирования об этом в канале команды в маттермосте.

## Описание

Утилита работает следующим образом. В конфигурации задается команды и список участников.
При вызове утилиты без параметров расчет происходит на текущую дату.
По указанной дате утилита определяет тип дня (рабочий | сокращенный | выходной).
Далее по каждому пользователя утилита находит все задачи, в которых пользователь проводил списание времени.
Далее по всем задачам запрашиваются все записи и производится фильтрация по пользователю и дате, формируется список списаний.
По списку списаний производится расчет оставшегося времени к отработке за день для каждого пользователя.
По отчету формируется сообщение в канал команды с тегированием пользователей, которые не списали указанную дневную норму.
После того как отчет сформирован происходит отправка сообщения в канал команды в маттермосте.

## Требования

1. Развернутый сервис jira (не важно облачный или развернутая корпоративная версия) и токена доступа к сервису.
2. Развернутый сервис Mattermost и токен доступа к нему.
3. Наличие сетевого доступа к сервису isdayoff.ru (https, 443 порт).

## Сборка

```shell
CGO_ENABLED=0 go build
```

## Запуск

## Ключи запуска

Для вывода справки используйте команду:

```shell
./ts-notifier -h
```

### Файл конфигурации

По умолчанию сервис работает с файлом конфигурации с именем config.yml, расположенным в той же директории, отпуда производится запуск.
Если требуется изменить путь к конфигурационному файлу используйте аргумент `-c <путь к конфигурационному файлу>`

- Пример  
  ```shell
  ./ts-notifier -c /home/user/config.yml
  ```
Сервис при запуске будет искать конфигурационный файл по пути `/home/user/config.yml`.

### Отчетный день

По умолчанию при запуске сервис определяет текущий день по системному времени машины, на которой он запущен.
Если требуется указать произвольный день используйте аргумент `-d YYYY-MM-DD`

- Пример: 
  ```shell
  ./ts-notifier -d 2023-09-01
  ```

Сервис при запуске выполнит подсчет списанного времени за дату `2023-09-01`.

### Предварительная настройка

Перед запуском требуется произвести настройку. Скопируйте пример конфига из `config/config-example.yml` в текущий каталог и заполните его.
1. Укажите адрес и порт развернутого сервиса Jira в параметр `jira.url`.
2. Укажите токен для аутентификации в параметр `jira.auth_token`.
3. Укажите адрес и порт развернутого сервиса Mattermost в параметр `notifier.mattermost.url`.
4. Укажите токен для аутентификации в параметр `notifier.mattermost.auth_token`.
5. Сформируйте команду, обязательно укажите идентификатор канала, в который будет отправлено уведомление. 

Все шаги выполнены, можете выполнить тестовый запуск.

## Отладка

Для отладки работы утилиты можно заметить отправку уведомлений в маттермост выводом в stdout.
Для этого в конфигурационном файле в секции `notifier.mattermost.url` укажите пустую строку. 
Результат работы программы будет выведен в stdout терминала.

После успешной отладки не забудьте установить адрес и порт сервиса Mattermost в `notifier.mattermost.url`.

# Запросы на улучшение и сообщения об ошибках
Если у вас есть запросы на улучшение или вы обнаружили ошибку, создайте [Issue](https://github.com/duke0x/ts-notifier/issues) на GitHub.

# Лицензия

- [LICENSE](LICENSE)
