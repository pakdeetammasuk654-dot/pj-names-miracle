# ชื่อดี 4 มิติ for iOS

Native SwiftUI client for the existing Go ranking API.

## Requirements

- Xcode 26+
- iOS 16+
- No paid Apple Developer membership is required for Simulator builds.

## Run

1. Open `ChueDee.xcodeproj`.
2. Select an iPhone Simulator.
3. Run the `ChueDee` scheme.

The Debug and Release configurations use the production API at
`https://xn--b3cu8e7ah6h.com/api/v1`. Change `API_BASE_URL` in the project build
settings when a local or staging environment is needed.

## Architecture

- `App/`: app entry point
- `Core/Networking/`: API transport and contract validation
- `Features/Search/`: SwiftUI screen and MVVM state
- `UI/`: shared native presentation components

Numerology, พลังเงา, Kalakini, semantic search, and ranking remain in the Go
backend. The iOS app only owns input, UI state, and rendering.
