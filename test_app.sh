#!/bin/bash

echo "Testing Golem Application..."
echo "================================"

# Check if database was created
if [ -f ~/.golem/golem.db ]; then
    echo "✓ Database created successfully"

    # Check database tables
    tables=$(sqlite3 ~/.golem/golem.db ".tables" 2>/dev/null)
    echo "✓ Database tables: $tables"

    # Check for any saved preferences
    prefs=$(sqlite3 ~/.golem/golem.db "SELECT COUNT(*) FROM preferences" 2>/dev/null)
    echo "✓ Preferences count: $prefs"

    # Check for any request history
    history=$(sqlite3 ~/.golem/golem.db "SELECT COUNT(*) FROM request_history" 2>/dev/null)
    echo "✓ Request history count: $history"
else
    echo "✗ Database not found"
fi

echo ""
echo "Application Status:"
if pgrep -f golem > /dev/null; then
    echo "✓ Golem is running"
    echo "  PID: $(pgrep -f golem)"
else
    echo "✗ Golem is not running"
fi

echo ""
echo "You can now:"
echo "1. Enter a URL (e.g., https://jsonplaceholder.typicode.com/posts/1)"
echo "2. Select HTTP method (GET, POST, PUT, PATCH, DELETE)"
echo "3. Click Submit to make the request"
echo "4. View request history in the left panel"
echo "5. Click on any history item to reload it"
echo "6. Search history using the search bar"
echo "7. Export history to JSON file"