import SwiftUI

extension Color {
    static let brandBackground = Color(red: 0.027, green: 0.075, blue: 0.051) // #07130D
    static let brandGreen = Color(red: 0.043, green: 0.169, blue: 0.106)      // #0B2B1B
    static let brandCard = Color(red: 0.063, green: 0.208, blue: 0.133)       // #103522
    static let brandCardHover = Color(red: 0.090, green: 0.271, blue: 0.180)  // #17452E
    static let brandGold = Color(red: 0.831, green: 0.686, blue: 0.216)       // #D4AF37
    static let brandGoldLight = Color(red: 0.969, green: 0.910, blue: 0.659)  // #F7E8A8
    static let brandCream = Color(red: 1.0, green: 0.973, blue: 0.863)        // #FFF8DC
    static let brandRed = Color(red: 0.851, green: 0.325, blue: 0.310)        // #D9534F
    static let brandBubbleAdmin = Color(red: 0.290, green: 0.231, blue: 0.071) // #4A3B12
    static let brandBubbleCustomer = Color(red: 0.071, green: 0.200, blue: 0.133) // #123322
}

extension Font {
    static func sarabun(_ size: CGFloat, weight: Font.Weight = .regular) -> Font {
        let name: String
        switch weight {
        case .bold:
            name = "Sarabun-Bold"
        case .semibold:
            name = "Sarabun-SemiBold"
        case .medium:
            name = "Sarabun-Medium"
        default:
            name = "Sarabun-Regular"
        }
        return Font.custom(name, size: size)
    }
}
