# Changelog

## [0.2.0] - 2026-02-23

Added

- ARCHITECTURE.md documenting concurrency model and design decisions
- Comprehensive "What I Learned" section in README showcasing technical growth
- .env.example template for easy local setup
- Proper HTTP error responses for authentication failures
- Context propagation patterns throughout server code
- Structured logging with slog across all packages
  Changed
- Migrated CLI from basic flags to Kong for better command structure
- Updated Dockerfile to use multi-stage builds with minimal final image
- Improved error messages throughout client and server
- Replaced all fmt.Print statements with structured logging
- Enhanced README with clearer installation and deployment instructions
  Fixed
- Division by zero crash in win percentage calculation for new users
- Panic in main.go default case now exits gracefully with error message
- GitHub username now properly fetched via OAuth Device Flow
- WebSocket upgrade properly validates authentication before accepting connection
- Empty error messages in client backend now have descriptive context
  Security
- Added authentication validation to WebSocket upgrade handler
- Removed exposed credentials from docker-compose examples
- Improved session management with proper token handling

## [0.1.1] - 2026-02-13

### Added

- CHANGELOG file to show improvements and history of this project
- Updated demo.gif in README
