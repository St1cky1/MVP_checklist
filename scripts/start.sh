#!/bin/sh

echo "Running migrations..."
./migrate

echo "Seeding data..."
./seed

echo "Starting server..."
./main
