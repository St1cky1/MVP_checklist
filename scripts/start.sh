#!/bin/sh

# Ожидание доступности БД (базовый вариант)
echo "Waiting for database to be ready..."
sleep 5

echo "Running migrations..."
if ! ./migrate; then
  echo "Migrations failed!"
  exit 1
fi

echo "Seeding data..."
if ! ./seed; then
  echo "Seeding failed!"
  # Не выходим, если сид уже был сделан
fi

echo "Starting server..."
exec ./main
