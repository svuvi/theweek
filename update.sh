#!/bin/bash

SCRIPT_DIR=$(dirname "$0")

# 1. Загружаем обновление
echo "Загружаю обновление с github.com/svuvi/theweek.git..."
GIT_PULL_OUTPUT=$(git -C "$SCRIPT_DIR" pull origin main 2>&1)

if [ $? -ne 0 ]; then
    echo "Git pull не удался. Отмена."
    exit 1
fi

if echo "$GIT_PULL_OUTPUT" | grep -q "Already up to date"; then
    echo "Локальная копия уже обновлена. Отмена."
    exit 0
fi

# 2. Генерируем шаблоны
echo "Генерирую шаблоны templ generate..."
templ generate

if [ $? -ne 0 ]; then
    echo "templ generate не удался. Отмена."
    exit 1
fi

# 3. Билдим приложение
echo "Создаю билд приложения..."
go build -o "$SCRIPT_DIR/bin/theweek-new"

if [ $? -ne 0 ]; then
    echo "Билд не удался. Отмена."
    exit 1
fi

# 4. Останавливаем сервис
echo "Останавливаю сервис theweek..."
sudo service theweek stop

if [ $? -ne 0 ]; then
    echo "Не удалось остановить сервис theweek. Отмена."
    exit 1
fi

# 5. Бэкап базы данных
datetime=$(date +"%Y%m%d-%H%M%S")
backup_path="$HOME/DBbackups/$datetime-prod.db"
echo "Бэкаплю базу данных в $backup_path..."
cp "$SCRIPT_DIR/bin/database.db" "$backup_path"

if [ $? -ne 0 ]; then
    echo "Бэкап базы данных не удался. Восстанавливаю старую версию сервиса."
    sudo service theweek start
    if [ $? -ne 0 ]; then
        echo "Рестарт сервиса не удался. Требуется ручное вмешательство."
        exit 1
    fi
    exit 1
fi

# 6. Удаляем старую версию
echo "Удаляю старую версию..."
rm -f "$SCRIPT_DIR/bin/theweek"

if [ $? -ne 0 ]; then
    echo "Не удалось удалить старую версию. Отмена. Восстанавливаю старую версию сервиса."
    sudo service theweek start
    if [ $? -ne 0 ]; then
        echo "Рестарт сервиса не удался. Требуется ручное вмешательство."
        exit 1
    fi
    exit 1
fi

# 7. Переименовываем новую версию 
mv "$SCRIPT_DIR/bin/theweek-new" "$SCRIPT_DIR/bin/theweek"
if [ $? -ne 0 ]; then
    echo "Не удалось переименовать новый исполняемый файл. Отмена."
    exit 1
fi

# 8. Запускаем сервис
echo "Запускаю сервис theweek..."
sudo service theweek start

if [ $? -ne 0 ]; then
    echo "Рестарт сервиса не удался. Требуется ручное вмешательство."
    exit 1
fi

echo "Обновление и перезапуск завершены успешно"