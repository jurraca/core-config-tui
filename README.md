
# Bitcoin Core Config Generator

A terminal user interface (TUI) for generating Bitcoin Core configuration files, built with Go using [Charm's Huh](https://github.com/charmbracelet/huh) and [BubbleTea](https://github.com/charmbracelet/bubbletea) libraries.

## About

This tool provides an interactive, user-friendly way to create `bitcoin.conf` configuration files for Bitcoin Core. Instead of manually editing the configuration file, users can navigate through various options using an elegant terminal interface.

## Features

- Interactive form-based configuration
- Real-time validation of settings
- Conditional field display based on selections
- Status bar showing current selections

## Usage

Run the program:

```bash
go run main.go
```

Navigate through the form using:
- Arrow keys to move between fields
- Enter to select/confirm
- Space to toggle checkboxes
- Tab to move between sections

The generated configuration will be saved as `bitcoin.conf` in your current directory.

## Key Sections

- **Basics**: Core settings like data directory and network selection
- **RPCs**: RPC server configuration including auth and port settings
- **General Options**: Advanced settings for Bitcoin Core operation

## Development

Built using:
- [Huh](https://github.com/charmbracelet/huh) for form handling
- [BubbleTea](https://github.com/charmbracelet/bubbletea) for terminal UI
- [Lipgloss](https://github.com/charmbracelet/lipgloss) for styling

## License

This project is open source and available under the MIT License.
