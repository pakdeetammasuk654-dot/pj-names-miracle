import SwiftUI

struct PairCircles: View {
    let pairs: [(value: String, isBad: Bool)]
    var size: CGFloat = 36
    var fontSize: CGFloat = 13
    var lineWidth: CGFloat = 1.8
    var spacing: CGFloat = 6
    var onPairTap: ((String) -> Void)? = nil

    var body: some View {
        HStack(spacing: spacing) {
            ForEach(Array(pairs.enumerated()), id: \.offset) { _, pair in
                let trimmed = pair.value.trimmingCharacters(in: .whitespacesAndNewlines)
                if !trimmed.isEmpty {
                    let status = pair.isBad ? "ควรพิจารณา" : "ส่งเสริม"
                    Button {
                        onPairTap?(trimmed)
                    } label: {
                        Text(trimmed)
                            .font(.system(size: fontSize, weight: .bold, design: .rounded))
                            .foregroundStyle(.white)
                            .frame(width: size, height: size)
                            .background(pair.isBad ? Color.statusCaution : Color.statusPositive)
                            .clipShape(Circle())
                            .overlay(Circle().stroke(Color.brandGold, lineWidth: lineWidth))
                    }
                    .buttonStyle(.plain)
                    .allowsHitTesting(onPairTap != nil)
                    .accessibilityLabel("คู่เลข \(trimmed): \(status)")
                    .accessibilityHint(onPairTap == nil ? "" : "แตะเพื่อดูความหมาย")
                }
            }
        }
    }
}
