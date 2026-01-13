#!/bin/bash

# Register Tic Tac Toe app in Identity Service database

IDENTITY_DB="/home/andrew/pubgames-v2/identity-service/data/identity.db"

echo "Registering Tic Tac Toe in Identity Service..."

sqlite3 "$IDENTITY_DB" <<SQL
INSERT OR REPLACE INTO apps (id, name, url, description, icon, is_active)
VALUES (4, 'Tic Tac Toe', 'http://localhost:30040', 'Multiplayer Tic Tac Toe game', '📤', 1);
SQL

if [ $? -eq 0 ]; then
    echo "✅ Tic Tac Toe registered successfully!"
    echo ""
    echo "The app should now appear in the Identity Service app launcher."
    echo ""
    echo "To start the app:"
    echo "1. cd /home/andrew/pubgames-v2/tic-tac-toe"
    echo "2. Terminal 1: go run *.go"
    echo "3. Terminal 2: npm start"
    echo ""
    echo "Access at: http://localhost:30040"
else
    echo "❌ Failed to register app"
    exit 1
fi
