import SwiftUI

struct RadarWeightSelector: View {
    @Binding var weights: RankingWeights
    @State private var isPulsing = false

    var body: some View {
        GeometryReader { geometry in
            let size = geometry.size
            // Keep a dedicated gutter for the four corner icons.
            let top = CGPoint(x: size.width / 2, y: 28)
            let left = CGPoint(x: 28, y: size.height / 2)
            let right = CGPoint(x: size.width - 28, y: size.height / 2)
            let bottom = CGPoint(x: size.width / 2, y: size.height - 28)
            let numerologyPoint = CGPoint(
                x: top.x * CGFloat(weights.numerology),
                y: top.y * CGFloat(weights.numerology)
            )
            let semanticPoint = CGPoint(
                x: left.x * CGFloat(weights.semantic),
                y: left.y * CGFloat(weights.semantic)
            )
            let soundPoint = CGPoint(
                x: bottom.x * CGFloat(weights.sound),
                y: bottom.y * CGFloat(weights.sound)
            )
            let similarityPoint = CGPoint(
                x: right.x * CGFloat(weights.similarity),
                y: right.y * CGFloat(weights.similarity)
            )
            let pointX = numerologyPoint.x + semanticPoint.x + soundPoint.x + similarityPoint.x
            let pointY = numerologyPoint.y + semanticPoint.y + soundPoint.y + similarityPoint.y
            let point = CGPoint(x: pointX, y: pointY)

            ZStack {
                Path { path in
                    path.move(to: top); path.addLine(to: right); path.addLine(to: bottom); path.addLine(to: left); path.closeSubpath()
                }
                .fill(LinearGradient(colors: [.brandGold.opacity(0.34), .brandGreen.opacity(0.5)], startPoint: .top, endPoint: .bottom))
                .overlay {
                    Path { path in
                        path.move(to: top); path.addLine(to: right); path.addLine(to: bottom); path.addLine(to: left); path.closeSubpath()
                    }.stroke(LinearGradient(colors: [.brandGold, .green.opacity(0.8)], startPoint: .leading, endPoint: .trailing), lineWidth: 2)
                }
                Path { path in
                    path.move(to: top); path.addLine(to: bottom)
                    path.move(to: left); path.addLine(to: right)
                }
                .stroke(.white.opacity(0.22), lineWidth: 1)
                Text("⭐").position(x: top.x, y: 10)
                Text("🎯").position(x: 10, y: left.y)
                Text("🔍").position(x: size.width - 10, y: right.y)
                Text("🎶").position(x: bottom.x, y: size.height - 10)
                Circle()
                    .stroke(Color.brandGold.opacity(isPulsing ? 0.04 : 0.48), lineWidth: 2)
                    .frame(width: isPulsing ? 34 : 24, height: isPulsing ? 34 : 24)
                    .position(point)
                Circle()
                    .fill(
                        RadialGradient(
                            colors: [
                                Color(red: 1.0, green: 0.97, blue: 0.78).opacity(isPulsing ? 0.80 : 0.94),
                                Color.brandGold.opacity(isPulsing ? 0.30 : 0.58),
                                .clear
                            ],
                            center: .center,
                            startRadius: 1,
                            endRadius: 22
                        )
                    )
                    .frame(width: isPulsing ? 44 : 34, height: isPulsing ? 44 : 34)
                    .position(point)
                Circle()
                    .fill(.white)
                    .frame(width: 15, height: 15)
                    .overlay(Circle().fill(Color.orange).frame(width: 8, height: 8))
                    .overlay(Circle().fill(Color.white.opacity(0.95)).frame(width: 3.6, height: 3.6))
                    .shadow(color: Color.brandGold.opacity(isPulsing ? 0.48 : 0.78), radius: isPulsing ? 9 : 5)
                    .position(point)
            }
            .contentShape(Rectangle())
            .gesture(DragGesture(minimumDistance: 0).onChanged { update($0.location, in: size) })
        }
        .frame(height: 145)
        .accessibilityLabel("ปรับสัดส่วนสูตรจัดอันดับชื่อดี 4 มิติ")
        .onAppear {
            guard !isPulsing else { return }
            withAnimation(.easeInOut(duration: 1.05).repeatForever(autoreverses: true)) {
                isPulsing = true
            }
        }
    }

    private func update(_ point: CGPoint, in size: CGSize) {
        let center = CGPoint(x: size.width / 2, y: size.height / 2)
        let radiusX = size.width / 2 - 28
        let radiusY = size.height / 2 - 28
        guard radiusX > 0, radiusY > 0 else { return }

        var x = Double((point.x - center.x) / radiusX)
        var y = Double((point.y - center.y) / radiusY)
        let scale = abs(x) + abs(y)
        if scale > 1 {
            x /= scale
            y /= scale
        }

        let u = (x + y + 1) / 2
        let v = (x - y + 1) / 2
        let raw = [
            (1 - u) * v,
            (1 - u) * (1 - v),
            u * (1 - v),
            u * v
        ]
        let quantized = quantize(raw)
        weights = RankingWeights(
            numerology: quantized[0],
            semantic: quantized[1],
            sound: quantized[2],
            similarity: quantized[3]
        )
    }

    private func quantize(_ values: [Double]) -> [Double] {
        let rawUnits = values.map { $0 * 20 }
        var units = rawUnits.map { Int(floor($0)) }
        var remaining = 20 - units.reduce(0, +)
        let fractions = rawUnits.enumerated().sorted { lhs, rhs in
            lhs.element - floor(lhs.element) > rhs.element - floor(rhs.element)
        }
        for entry in fractions where remaining > 0 {
            units[entry.offset] += 1
            remaining -= 1
        }
        return units.map { Double($0) / 20 }
    }
}
