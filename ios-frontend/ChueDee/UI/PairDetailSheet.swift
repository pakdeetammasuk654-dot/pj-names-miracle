import SwiftUI

struct PairDetailSheet: View {
    let detail: PairDetailResponse?
    let isLoading: Bool
    let errorMessage: String?
    let onDismiss: () -> Void

    var body: some View {
        ScrollView {
            VStack(spacing: 16) {
                if isLoading && detail == nil {
                    ProgressView()
                        .tint(.brandGold)
                        .controlSize(.large)
                        .frame(maxWidth: .infinity, minHeight: 180)
                } else if let detail {
                    detailContent(detail)
                } else {
                    VStack(spacing: 12) {
                        Image(systemName: "exclamationmark.triangle.fill")
                            .font(.system(size: 30))
                            .foregroundStyle(Color.brandGold)
                        Text(errorMessage ?? "ไม่สามารถโหลดความหมายคู่เลขได้")
                            .font(.sarabun(14, weight: .semibold))
                            .multilineTextAlignment(.center)
                            .foregroundStyle(Color.brandCream)
                    }
                    .frame(maxWidth: .infinity, minHeight: 180)
                }
            }
            .padding(.horizontal, 24)
            .padding(.vertical, 12)
        }
        .background(Color.brandBackground.ignoresSafeArea())
        .preferredColorScheme(.dark)
    }

    @ViewBuilder
    private func detailContent(_ detail: PairDetailResponse) -> some View {
        HStack(spacing: 10) {
            Text("คู่เลข")
                .font(.sarabun(18, weight: .bold))
                .foregroundStyle(Color.brandCream)
            PairCircles(pairs: [(detail.pairNumber, !detail.isGood)])
            Spacer()
            Button(action: onDismiss) {
                Image(systemName: "xmark")
                    .font(.system(size: 17, weight: .semibold))
                    .foregroundStyle(Color.brandCream.opacity(0.82))
                    .frame(width: 40, height: 40)
            }
            .buttonStyle(.plain)
            .accessibilityLabel("ปิด")
        }

        Divider().overlay(Color.brandGold.opacity(0.42))

        if !detail.miracleDesc.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            HStack(spacing: 10) {
                Image(systemName: "sparkles")
                    .font(.system(size: 19, weight: .semibold))
                    .foregroundStyle(Color.brandGold)
                Text(detail.miracleDesc)
                    .font(.sarabun(14, weight: .bold))
                    .foregroundStyle(Color.brandCream)
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
            .padding(14)
            .background(Color.brandGreen.opacity(0.68))
            .clipShape(RoundedRectangle(cornerRadius: 14))
            .overlay(RoundedRectangle(cornerRadius: 14).stroke(Color.brandGold.opacity(0.34)))
        }

        Text(detailText(detail))
            .font(.sarabun(13))
            .foregroundStyle(Color.brandCream.opacity(0.92))
            .lineSpacing(6)
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(16)
            .background(Color(red: 0.06, green: 0.15, blue: 0.11))
            .clipShape(RoundedRectangle(cornerRadius: 14))
            .overlay(RoundedRectangle(cornerRadius: 14).stroke(Color.brandGold.opacity(0.24)))
    }

    private func detailText(_ detail: PairDetailResponse) -> String {
        let full = detail.miracleDetail.trimmingCharacters(in: .whitespacesAndNewlines)
        if !full.isEmpty { return full }
        let short = detail.miracleDesc.trimmingCharacters(in: .whitespacesAndNewlines)
        if !short.isEmpty { return short }
        return "คู่เลข \(detail.pairNumber) เป็นคู่เลขตามตำราเลขศาสตร์ไทย"
    }
}
