# Golem

A lightweight, cross-platform API testing application written in Go, inspired by Postman. Golem provides a clean GUI interface for testing HTTP endpoints with persistent request history and collections support.

## Features

- **HTTP Methods Support**: GET, POST, PUT, PATCH, DELETE
- **Request History**: Automatically saves all requests with responses
- **Search Functionality**: Search through request history by URL, method, or status code
- **Collections**: Organize your saved requests into collections
- **Persistent Storage**: SQLite database for reliable data persistence
- **Export/Import**: Export your request history to JSON for backup or sharing
- **Modern GUI**: Built with the Fyne framework for a native cross-platform experience
- **Lightweight**: Single binary with minimal dependencies
- **Fast**: Written in Go for optimal performance

## Prerequisites

- Go 1.19 or higher
- Git (for cloning the repository)

## Building from Source

### Clone the Repository

```bash
git clone https://github.com/wickeddoc/golem.git
cd golem
```

### Install Dependencies

```bash
go mod download
```

### Build the Binary

```bash
go build -o golem .
```

This will create a `golem` executable in your current directory.

### Optional: Install System-wide

```bash
sudo cp golem /usr/local/bin/
```

Or on Windows, add the binary location to your PATH.

## Running the Application

### From Source
```bash
go run main.go
```

### Using Built Binary
```bash
./golem
```

## Usage

1. **Making Requests**
   - Enter the URL in the URL field
   - Select the HTTP method from the dropdown
   - Click "Submit" to execute the request
   - View the response in the main panel

2. **Request History**
   - All requests are automatically saved to history
   - Click on any history item in the left panel to reload it
   - Use the search bar to filter history
   - Export history to JSON for backup

3. **Keyboard Shortcuts**
   - `Ctrl+Enter`: Submit request
   - `F6`: Focus URL field

## Data Storage

Golem stores all data in a SQLite database located at:
- **Linux/macOS**: `~/.golem/golem.db`
- **Windows**: `%USERPROFILE%\.golem\golem.db`

The database includes:
- Application preferences (window size, last used URL/method)
- Complete request history
- Saved request collections
- Request templates

## Project Structure

```
golem/
├── main.go           # Application entry point and core logic
├── storage/
│   ├── db.go        # Database initialization and connection management
│   └── models.go    # Data models and CRUD operations
├── ui/
│   └── history.go   # History panel UI component
├── go.mod           # Go module dependencies
└── go.sum           # Dependency checksums
```

## Dependencies

- **[Fyne](https://fyne.io/)** v2.6.2 - Cross-platform GUI framework
- **[modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite)** - Pure Go SQLite driver

## Development

### Updating Dependencies
```bash
go mod tidy
go mod download
```

### Building for Different Platforms

**Linux:**
```bash
GOOS=linux GOARCH=amd64 go build -o golem-linux .
```

**Windows:**
```bash
GOOS=windows GOARCH=amd64 go build -o golem.exe .
```

**macOS:**
```bash
GOOS=darwin GOARCH=amd64 go build -o golem-mac .
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## Roadmap

- [ ] Request body support (JSON, form data, raw text)
- [ ] Custom headers management
- [ ] Environment variables
- [ ] Response syntax highlighting
- [ ] Request authentication (Basic, Bearer, API Key)
- [ ] WebSocket support
- [ ] Response time graphs
- [ ] Import/Export Postman collections
- [ ] Dark/Light theme toggle
- [ ] Request chaining
- [ ] Pre-request scripts
- [ ] Response tests/assertions

## License

This project is open source and available under the [MIT License](LICENSE).

## Acknowledgments

- Inspired by [Postman](https://www.postman.com/)
- Built with [Fyne](https://fyne.io/) framework
- Uses [SQLite](https://www.sqlite.org/) for data persistence

## Contact

For issues, questions, or suggestions, please open an issue on GitHub.