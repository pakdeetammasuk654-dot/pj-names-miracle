import SwiftUI

struct MagicalLoadingView: View {
    private let messages = [
        "✨ กำลังเปิดม่านรายชื่อมงคล...",
        "🔮 กำลังร้อยเรียงพลังของตัวเลข...",
        "🌟 กำลังค้นหาชื่อที่เหมาะกับคุณ...",
        "📜 คำนวณเลขศาสตร์และพลังเงา..."
    ]
    @State private var messageIndex = 0

    var body: some View {
        TimelineView(.animation(minimumInterval: 1.0 / 30.0)) { timeline in
            let time = timeline.date.timeIntervalSinceReferenceDate
            ZStack {
                Color.brandBackground.ignoresSafeArea()
                particleField(time: time)
                RadialGradient(colors: [Color.brandGold.opacity(0.15), Color(hex: "#26160E").opacity(0.25), .clear], center: .top, startRadius: 10, endRadius: 330).ignoresSafeArea()
                VStack(spacing: 14) {
                    statusBanner(time: time)
                    skeletonHeader(time: time)
                    ForEach(0..<5, id: \.self) { index in skeletonRow(rank: index + 1, time: time, highlighted: index == 0) }
                    Spacer(minLength: 0)
                }
                .padding(.horizontal, 16).padding(.top, 18)
            }
        }
        .task {
            while !Task.isCancelled {
                try? await Task.sleep(for: .seconds(1.4))
                guard !Task.isCancelled else { return }
                withAnimation(.easeInOut(duration: 0.4)) { messageIndex = (messageIndex + 1) % messages.count }
            }
        }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(messages[messageIndex])
    }

    private func statusBanner(time: TimeInterval) -> some View {
        VStack(spacing: 13) {
            HStack(spacing: 12) {
                ZStack {
                    Image(systemName: "sparkles").font(.system(size: 29, weight: .bold)).foregroundStyle(Color.brandGold).rotationEffect(.degrees(time * 90)).scaleEffect(0.98 + sin(time * 4) * 0.12)
                    Image(systemName: "star.fill").font(.system(size: 12)).foregroundStyle(Color(hex: "#FFF0A3")).rotationEffect(.degrees(-time * 135))
                }
                Text(messages[messageIndex]).id(messageIndex).font(.sarabun(15, weight: .bold)).transition(.opacity.combined(with: .move(edge: .bottom)))
            }.frame(maxWidth: .infinity)
            GeometryReader { proxy in
                let travel = (CGFloat(time.truncatingRemainder(dividingBy: 1.5)) / 1.5) * (proxy.size.width + 100) - 100
                Capsule().fill(Color.white.opacity(0.12)).overlay(alignment: .leading) {
                    LinearGradient(colors: [.brandGold, Color(hex: "#FFF0A3"), Color(hex: "#FFDF85"), Color(hex: "#A37C10"), .brandGold], startPoint: .leading, endPoint: .trailing).frame(width: 110).offset(x: travel)
                }.clipShape(Capsule())
            }.frame(height: 4)
        }
        .padding(16).background(LinearGradient(colors: [Color(hex: "#26160E"), Color(hex: "#160C08"), Color(hex: "#382014")], startPoint: .leading, endPoint: .trailing))
        .clipShape(RoundedRectangle(cornerRadius: 18)).overlay(RoundedRectangle(cornerRadius: 18).stroke(LinearGradient(colors: [.brandGold, Color(hex: "#FFF0A3"), .brandGold], startPoint: .leading, endPoint: .trailing), lineWidth: 1.5))
    }

    private func skeletonHeader(time: TimeInterval) -> some View {
        HStack { shimmer(width: 28, time: time); shimmer(width: 110, time: time); Spacer(); shimmer(width: 72, time: time); shimmer(width: 48, time: time) }
            .padding(.horizontal, 13).frame(height: 35).background(Color(hex: "#1F120B")).clipShape(RoundedRectangle(cornerRadius: 8))
    }

    private func skeletonRow(rank: Int, time: TimeInterval, highlighted: Bool) -> some View {
        HStack(spacing: 12) {
            Text("#\(rank)").font(.system(size: 14, weight: .bold, design: .rounded)).foregroundStyle(highlighted ? Color.brandGold : .secondary).frame(width: 34)
            VStack(alignment: .leading, spacing: 7) { shimmer(width: CGFloat(95 + rank * 8), time: time); shimmer(width: 120, height: 8, time: time) }
            Spacer(); Circle().fill(shimmerGradient(time: time)).frame(width: 40, height: 40); Circle().fill(shimmerGradient(time: time + 0.2)).frame(width: 40, height: 40)
        }
        .padding(.horizontal, 13).frame(height: 74).background(highlighted ? Color(hex: "#26160E") : Color(hex: "#160C08")).clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private func shimmer(width: CGFloat, height: CGFloat = 12, time: TimeInterval) -> some View {
        RoundedRectangle(cornerRadius: height / 2).fill(shimmerGradient(time: time)).frame(width: width, height: height)
    }

    private func shimmerGradient(time: TimeInterval) -> LinearGradient {
        let shift = CGFloat((sin(time * 3) + 1) / 2)
        return LinearGradient(colors: [Color.white.opacity(0.06), Color.brandGold.opacity(0.25), Color(hex: "#FFF0A3").opacity(0.55), Color.brandGold.opacity(0.25), Color.white.opacity(0.06)], startPoint: UnitPoint(x: shift - 0.5, y: 0), endPoint: UnitPoint(x: shift + 0.5, y: 1))
    }

    private func particleField(time: TimeInterval) -> some View {
        GeometryReader { proxy in
            ForEach(0..<16, id: \.self) { index in
                let x = (CGFloat(index * 73 % 100) / 100) * proxy.size.width
                let cycle = (CGFloat(time * 18) + CGFloat(index * 61)).truncatingRemainder(dividingBy: proxy.size.height + 80)
                Circle().fill(index.isMultiple(of: 2) ? Color.brandGold : Color(hex: "#FFF0A3")).frame(width: CGFloat(2 + index % 4), height: CGFloat(2 + index % 4)).opacity(0.18 + 0.22 * sin(time * 2 + Double(index))).position(x: x, y: proxy.size.height + 30 - cycle)
            }
        }.ignoresSafeArea().allowsHitTesting(false)
    }
}
