import SwiftUI

struct AuspiciousCalendarView: View {
    @StateObject private var viewModel = AuspiciousViewModel()
    @State private var startDate = ThaiDateCodec.shared.today
    @State private var endDate = ThaiDateCodec.shared.addDays(90, to: ThaiDateCodec.shared.today)
    @State private var birthDate = ThaiDateCodec.shared.today
    @State private var birthTime = Date()
    @State private var hasBirthDate = false
    @State private var hasBirthTime = false
    @State private var preferWeekend = false
    @State private var avoidWanPhra = false
    @State private var monkCeremony = false
    @State private var formError: String?
    @State private var showsResults = false
    @State private var showsChatSheet = false
    @State private var readyButtonPulse = false

    private let dateCodec = ThaiDateCodec.shared

    var body: some View {
        ZStack {
            Color.brandBackground.ignoresSafeArea()
            ScrollViewReader { proxy in
                ScrollView { auspiciousScrollContent }
                .onChange(of: hasBirthTime) { _, isSet in
                    guard isSet else { return }
                    withAnimation(.easeInOut(duration: 0.45)) {
                        proxy.scrollTo("birth-time-field", anchor: .top)
                    }
                }
            }
        }
        .navigationTitle("ดูฤกษ์ยาม")
        .navigationBarTitleDisplayMode(.inline)
        .toolbarBackground(Color.brandGreen, for: .navigationBar)
        .toolbarBackground(.visible, for: .navigationBar)
        .font(.sarabun(14))
        .tint(.brandGold)
        .preferredColorScheme(.dark)
        .floatingChatOverlay(showsChatSheet: $showsChatSheet)
        .task { await viewModel.loadActivities() }
        .onChange(of: viewModel.selectedActivityCode) { _, code in
            if code == "ordination" { monkCeremony = true }
            formError = nil
        }
        .onChange(of: hasBirthDate) { formError = nil; refreshTaksa() }
        .onChange(of: hasBirthTime) { formError = nil; refreshTaksa() }
        .onChange(of: birthDate) { refreshTaksa() }
        .onChange(of: birthTime) { refreshTaksa() }
        .onAppear { readyButtonPulse = isReadyToSearch }
        .onChange(of: isReadyToSearch) { _, ready in readyButtonPulse = ready }
        .sheet(isPresented: $showsResults) {
            if let result = viewModel.result {
                AuspiciousResultsView(result: result)
                    .presentationDetents([.large])
                    .presentationDragIndicator(.visible)
            }
        }
    }

    private var auspiciousScrollContent: some View {
        LazyVStack(spacing: 18) {
            stepHeading(1, "คุณกำลังจะทำอะไร?", "เลือกหมวด แล้วเลือกเหตุการณ์สำคัญ")
            activityPicker

            stepHeading(2, "ช่วงวันที่สะดวก", "ระบบจะเปรียบเทียบได้สูงสุด 366 วัน")
            dateRangeCard

            stepHeading(3, "ปรับให้เหมาะกับคุณ", "วันเกิดช่วยให้ระบบผสานทักษาจรเฉพาะบุคคล")
            birthCard
            taksaPanel

            stepHeading(4, "ความสะดวกและรูปแบบพิธี", "ตัวเลือกของคุณแยกจากกฎตำราอย่างชัดเจน")
            preferenceCard

            if let message = formError ?? viewModel.errorMessage {
                errorCard(message)
            }
            searchButton
        }
        .padding(16)
        .padding(.bottom, 30)
    }

    private var activityPicker: some View {
        Group {
            if viewModel.isLoadingActivities {
                HStack(spacing: 10) {
                    ProgressView().tint(.brandGold)
                    Text("กำลังโหลดรายการกิจกรรม…").foregroundStyle(Color.brandCream.opacity(0.72))
                }
                .frame(maxWidth: .infinity, minHeight: 86)
                .luxuryCard()
            } else {
                Menu {
                    ForEach(viewModel.activities) { activity in
                        Button("\(activity.icon) \(activity.nameTH)") {
                            viewModel.selectActivity(activity.code)
                            formError = nil
                        }
                    }
                } label: {
                    HStack(spacing: 12) {
                        Text(viewModel.selectedActivity?.icon ?? "✦").font(.system(size: 28))
                        VStack(alignment: .leading, spacing: 3) {
                            Text("กิจกรรมที่เลือก").font(.sarabun(11)).foregroundStyle(Color.brandCream.opacity(0.65))
                            Text(viewModel.selectedActivity?.nameTH ?? "เลือกกิจกรรม")
                                .font(.sarabun(16, weight: .bold))
                                .foregroundStyle(Color.brandCream)
                                .multilineTextAlignment(.leading)
                        }
                        Spacer()
                        Image(systemName: "chevron.down").foregroundStyle(Color.brandGold)
                    }
                    .contentShape(Rectangle())
                }
                .buttonStyle(.plain)
                .luxuryCard()
            }
        }
    }

    private var dateRangeCard: some View {
        HStack(spacing: 8) {
            ThaiDateField(
                label: "วันเริ่มต้น",
                selection: $startDate,
                range: dateCodec.addDays(-365, to: endDate)...endDate
            )
            ThaiDateField(
                label: "วันสิ้นสุด",
                selection: $endDate,
                range: startDate...dateCodec.addDays(365, to: startDate)
            )
        }
        .luxuryCard()
    }

    private var birthCard: some View {
        VStack(alignment: .leading, spacing: 10) {
            ThaiDateField(
                label: "วันเกิด (พ.ศ. · จำเป็นสำหรับทักษาจร)",
                selection: $birthDate,
                range: dateCodec.date(year: 1900, month: 1, day: 1)...dateCodec.today,
                isSet: $hasBirthDate,
                placeholder: "แตะเพื่อเลือกวันเกิด",
                startWithYearSelection: true
            )
            if !hasBirthDate {
                Text("แตะช่องวันเกิด แล้วเลือกวันที่จากปฏิทิน พ.ศ.")
                    .font(.sarabun(11))
                    .foregroundStyle(Color.brandCream.opacity(0.65))
                    .padding(.leading, 2)
            }
            Divider().overlay(Color.brandGold.opacity(0.25))
            ThaiTimeField(selection: $birthTime, isSet: $hasBirthTime)
                .id("birth-time-field")
            if !hasBirthTime {
                Text("จำเป็นต้องระบุเวลาเกิด เพื่อคำนวณทักษาจรโดยใช้เวลาเปลี่ยนวัน 06:00 น.")
                    .font(.sarabun(11))
                    .foregroundStyle(Color.brandGold)
                    .padding(.leading, 2)
            }
        }
        .luxuryCard()
    }

    @ViewBuilder private var taksaPanel: some View {
        if viewModel.isLoadingTaksa {
            HStack(spacing: 10) {
                ProgressView().tint(.brandGold)
                Text("กำลังคำนวณทักษาจร…").foregroundStyle(Color.brandCream.opacity(0.7))
            }
            .frame(maxWidth: .infinity)
            .luxuryCard()
        } else if let taksa = viewModel.taksa {
            TaksaWheelPanelView(result: taksa)
        } else if let error = viewModel.taksaErrorMessage {
            Text(error).foregroundStyle(Color.orange).luxuryCard()
        }
    }

    private var preferenceCard: some View {
        VStack(spacing: 8) {
            preferenceToggle("เน้นวันเสาร์–อาทิตย์", "เพิ่มน้ำหนักด้านความสะดวก", isOn: $preferWeekend)
            Divider().overlay(Color.brandGold.opacity(0.18))
            preferenceToggle("หลีกเลี่ยงวันพระ", "ตัดวันพระออกจากผลลัพธ์", isOn: $avoidWanPhra)
            Divider().overlay(Color.brandGold.opacity(0.18))
            preferenceToggle("มีพิธีสงฆ์", "ตรวจข้อห้ามที่เกี่ยวกับพิธีสงฆ์", isOn: $monkCeremony)
        }
        .luxuryCard()
    }

    private var searchButton: some View {
        let missingMessage = missingFormMessage
        let isReady = isReadyToSearch
        let isMuted = missingMessage != nil || viewModel.isLoadingActivities
        return Button {
            Task { await submit() }
        } label: {
            HStack(spacing: 10) {
                if viewModel.isSearching {
                    ProgressView().tint(Color(hex: "#140B07"))
                    Text("กำลังเปรียบเทียบทุกวัน…")
                        .font(.sarabun(16, weight: .bold))
                } else if let missing = missingMessage {
                    Text(missing)
                        .font(.sarabun(15, weight: .heavy))
                        .kerning(0.5)
                } else {
                    Image(systemName: "calendar.badge.checkmark")
                    Text("ค้นหาวันที่เหมาะที่สุด")
                        .font(.sarabun(17, weight: .bold))
                }
            }
            .foregroundStyle(isMuted ? Color(white: 0.25) : Color(hex: "#140B07"))
            .frame(maxWidth: .infinity, minHeight: 54)
            .background(isMuted ? Color(white: 0.82) : Color.brandGold)
            .clipShape(RoundedRectangle(cornerRadius: 18))
        }
        .scaleEffect(isReady && readyButtonPulse ? 1.025 : 1)
        .shadow(
            color: isReady ? Color.brandGold.opacity(readyButtonPulse ? 0.60 : 0.24) : .clear,
            radius: isReady && readyButtonPulse ? 14 : 0,
            y: isReady && readyButtonPulse ? 5 : 0
        )
        .animation(
            isReady ? .easeInOut(duration: 0.9).repeatForever(autoreverses: true) : .default,
            value: readyButtonPulse
        )
        .disabled(viewModel.isSearching || viewModel.isLoadingActivities || missingMessage != nil)
    }

    private var isReadyToSearch: Bool {
        missingFormMessage == nil && !viewModel.isSearching
    }

    private var missingFormMessage: String? {
        if viewModel.isLoadingActivities { return "กำลังโหลดรายการกิจกรรม…" }
        if viewModel.selectedActivityCode.isEmpty { return "⚠️ โปรดเลือกกิจกรรม" }
        if !hasBirthDate { return "⚠️ โปรดกรอกวันเกิด" }
        if !hasBirthTime { return "⚠️ โปรดกรอกเวลาเกิด" }
        return nil
    }

    private func submit() async {
        formError = nil
        guard !viewModel.selectedActivityCode.isEmpty else { formError = "กรุณาเลือกกิจกรรมที่ต้องการหาฤกษ์"; return }
        guard hasBirthDate else { formError = "กรุณาระบุวันเกิดเพื่อคำนวณทักษาจร"; return }
        guard hasBirthTime else { formError = "กรุณาระบุเวลาเกิดเพื่อคำนวณทักษาจร"; return }
        guard endDate >= startDate else { formError = "วันที่สิ้นสุดต้องไม่มาก่อนวันที่เริ่มต้น"; return }
        guard dateCodec.days(from: startDate, to: endDate) <= 365 else { formError = "เลือกช่วงเวลาได้ไม่เกิน 366 วัน"; return }

        await viewModel.search(
            AuspiciousSearchRequest(
                activityCode: viewModel.selectedActivityCode,
                startDate: dateCodec.apiDate(startDate),
                endDate: dateCodec.apiDate(endDate),
                birthDate: dateCodec.apiDate(birthDate),
                birthTime: dateCodec.apiTime(birthTime),
                preferWeekend: preferWeekend,
                avoidWanPhra: avoidWanPhra,
                monkCeremony: monkCeremony
            )
        )
        showsResults = viewModel.result != nil
    }

    private func refreshTaksa() {
        guard hasBirthDate, hasBirthTime else { return }
        viewModel.loadTaksa(
            birthDate: dateCodec.apiDate(birthDate),
            birthTime: dateCodec.apiTime(birthTime)
        )
    }

    private func stepHeading(_ number: Int, _ title: String, _ subtitle: String) -> some View {
        HStack(spacing: 11) {
            Text("\(number)")
                .font(.sarabun(17, weight: .bold))
                .foregroundStyle(Color.brandGreen)
                .frame(width: 42, height: 42)
                .background(Color.brandGold)
                .clipShape(Circle())
            VStack(alignment: .leading, spacing: 2) {
                Text(title).font(.prompt(19, weight: .bold)).foregroundStyle(Color.brandCream)
                Text(subtitle).font(.sarabun(12)).foregroundStyle(Color.brandCream.opacity(0.65))
            }
            Spacer()
        }
    }

    private func preferenceToggle(_ title: String, _ subtitle: String, isOn: Binding<Bool>) -> some View {
        Toggle(isOn: isOn) {
            VStack(alignment: .leading, spacing: 2) {
                Text(title).foregroundStyle(Color.brandCream)
                Text(subtitle).font(.sarabun(11)).foregroundStyle(Color.brandCream.opacity(0.65))
            }
        }
        .tint(.brandGold)
    }

    private func errorCard(_ message: String) -> some View {
        HStack(alignment: .top, spacing: 9) {
            Image(systemName: "exclamationmark.triangle.fill")
            Text(message).frame(maxWidth: .infinity, alignment: .leading)
            Button { formError = nil; viewModel.errorMessage = nil } label: { Image(systemName: "xmark") }
        }
        .foregroundStyle(Color(red: 1, green: 0.78, blue: 0.68))
        .padding(14)
        .background(Color.red.opacity(0.18))
        .clipShape(RoundedRectangle(cornerRadius: 14))
    }
}

private struct TaksaWheelPanelView: View {
    let result: TaksaCalculationResponse
    @State private var selectedRole: TaksaRole? = nil

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("กระดานทดสำหรับวิเคราะห์ฤกษ์ยาม")
                .font(.sarabun(11))
                .foregroundStyle(Color.brandCream.opacity(0.58))
            Text(result.isTodayDefault ? "✨ ทักษาพลังประจำวัน" : "🧭 ทักษาจรประจำปีของคุณ")
                .font(.sarabun(18, weight: .bold))
                .foregroundStyle(Color.brandGold)
            Text(summary)
                .font(.sarabun(12))
                .foregroundStyle(Color.brandCream.opacity(0.68))
            Text("แตะแต่ละช่องเพื่อดูความหมายต่อฤกษ์ของคุณ")
                .font(.sarabun(11, weight: .semibold))
                .foregroundStyle(Color.brandGold)
            TaksaWheelView(result: result) { role in
                selectedRole = role
            }
                .id("\(result.progressedAge)-\(result.landedPlanetCode)-\(result.wheelRotationDegrees)")

        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, 4)
        .padding(.vertical, 6)
        .sheet(item: $selectedRole) { role in
            TaksaRoleDetailSheet(role: role)
                .presentationDetents([.medium])
                .presentationDragIndicator(.visible)
        }
    }

    private var summary: String {
        if result.isTodayDefault { return "แสดงพลังของวันนี้ · ระบุวันเกิดเพื่อดูผลเฉพาะบุคคล" }
        if result.isCenterCell { return "เกิด\(result.birthDayName) · อายุย่าง \(result.progressedAge) ปี · ตกตากลาง ๙" }
        return "เกิด\(result.birthDayName) · อายุย่าง \(result.progressedAge) ปี · ตก\(result.landedPlanetName)"
    }
}

private struct TaksaWheelView: View {
    let result: TaksaCalculationResponse
    let onRoleSelected: (TaksaRole) -> Void
    @State private var animatedRotation = 0.0

    var body: some View {
        GeometryReader { geometry in
            let side = min(geometry.size.width, geometry.size.height)
            let center = CGPoint(x: geometry.size.width / 2, y: geometry.size.height / 2)
            let outerRadius = side / 2 - 9
            let innerRadius = outerRadius * 0.39

            ZStack {
                Circle()
                    .fill(Color(hex: "#140B07"))
                    .frame(width: outerRadius * 2 + 10, height: outerRadius * 2 + 10)
                Canvas { context, _ in
                    for (index, role) in result.allRoles.enumerated() {
                        var path = Path()
                        path.move(to: center)
                        path.addArc(
                            center: center,
                            radius: outerRadius,
                            startAngle: .degrees(-112.5 + Double(index) * 45),
                            endAngle: .degrees(-68.2 + Double(index) * 45),
                            clockwise: false
                        )
                        path.closeSubpath()
                        context.fill(path, with: .color(taksaSectorColor(role, index: index)))
                    }
                }
                .rotationEffect(.degrees(animatedRotation - result.wheelRotationDegrees))

                Circle()
                    .fill(Color(hex: "#140B07"))
                    .frame(width: innerRadius * 2, height: innerRadius * 2)
                    .overlay(Circle().stroke(Color.brandGold.opacity(0.72), lineWidth: 1.5))

                ForEach(Array(result.allRoles.enumerated()), id: \.element.id) { index, role in
                    let angle = (-90.0 + Double(index) * 45.0) * .pi / 180
                    VStack(spacing: 1) {
                        Text("\(role.symbol) \(role.wheelDayName)")
                            .font(.system(size: 9, weight: .bold))
                            .foregroundStyle(Color.white)
                        Text(role.roleName.replacingOccurrences(of: "จร", with: ""))
                            .font(.sarabun(9, weight: .bold))
                            .foregroundStyle(taksaRoleColor(role))
                    }
                    .fixedSize()
                    .position(
                        x: center.x + cos(angle) * outerRadius * 0.66,
                        y: center.y + sin(angle) * outerRadius * 0.66
                    )
                }

                VStack(spacing: 2) {
                    Text(result.isTodayDefault ? "พลังทักษา" : "อายุย่าง")
                        .font(.sarabun(10, weight: .bold)).foregroundStyle(Color.brandGold)
                    Text(result.isTodayDefault ? "วันนี้" : "\(result.progressedAge) ปี")
                        .font(.sarabun(18, weight: .bold)).foregroundStyle(Color.white)
                    Text(result.isCenterCell ? "ตากลาง ๙" : "\(result.landedPlanetSymbol) ดาว \(result.landedPlanetNum)")
                        .font(.sarabun(9, weight: .semibold)).foregroundStyle(Color(hex: "#FFDF85"))
                }
                .frame(width: innerRadius * 1.8, height: innerRadius * 1.8)

                TaksaPointerShape()
                    .fill(Color.brandGold)
                    .frame(width: 18, height: 17)
                    .position(x: center.x, y: 10)

                Circle()
                    .stroke(Color.brandGold, lineWidth: 2)
                    .frame(width: outerRadius * 2 + 6, height: outerRadius * 2 + 6)
            }
            .frame(width: geometry.size.width, height: geometry.size.height)
            .contentShape(Rectangle())
            .gesture(
                DragGesture(minimumDistance: 0)
                    .onEnded { value in
                        let dx = value.location.x - center.x
                        let dy = value.location.y - center.y
                        let distance = sqrt(dx * dx + dy * dy)
                        guard distance >= innerRadius,
                              distance <= outerRadius,
                              !result.allRoles.isEmpty else { return }
                        let degrees = atan2(dy, dx) * 180 / .pi
                        let normalized = (degrees + 112.5).truncatingRemainder(dividingBy: 360) + 360
                        let index = Int(normalized.truncatingRemainder(dividingBy: 360) / 45)
                        guard result.allRoles.indices.contains(index) else { return }
                        UIImpactFeedbackGenerator(style: .light).impactOccurred()
                        onRoleSelected(result.allRoles[index])
                    }
            )
        }
        .aspectRatio(1, contentMode: .fit)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(
            result.isTodayDefault
                ? "วงล้อทักษาพลังประจำวัน ตก\(result.landedPlanetName)"
                : "วงล้อทักษาจร อายุย่าง \(result.progressedAge) ปี ตก\(result.landedPlanetName)"
        )
        .onAppear {
            animatedRotation = result.wheelRotationDegrees - 1080
            withAnimation(.easeInOut(duration: 1.8)) {
                animatedRotation = result.wheelRotationDegrees
            }
        }
    }
}

private struct TaksaRoleDetailSheet: View {
    let role: TaksaRole

    private var accent: Color { taksaRoleColor(role) }
    private var isCaution: Bool { role.impact == "caution" }

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack(spacing: 12) {
                Text(role.symbol).font(.system(size: 34))
                VStack(alignment: .leading, spacing: 3) {
                    Text(role.roleName)
                        .font(.sarabun(22, weight: .bold))
                        .foregroundStyle(accent)
                    Text(role.dayName)
                        .font(.sarabun(14))
                        .foregroundStyle(Color.brandCream)
                }
            }

            Text(isCaution ? "ควรระวังในการเลือกฤกษ์" : "พลังสนับสนุนการเลือกฤกษ์")
                .font(.sarabun(14, weight: .bold))
                .foregroundStyle(isCaution ? Color(red: 1, green: 0.65, blue: 0.65) : Color(hex: "#FFDF85"))

            detailSection(
                title: "ความหมาย",
                text: meaningText
            )
            detailSection(
                title: "ความสัมพันธ์กับฤกษ์ของคุณ",
                text: guidanceText
            )
            Spacer(minLength: 0)
        }
        .padding(22)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color.brandBackground.ignoresSafeArea())
        .preferredColorScheme(.dark)
    }

    private var meaningText: String {
        guard let meaning = role.meaning, !meaning.isEmpty else { return role.description }
        return meaning
    }

    private var guidanceText: String {
        guard let guidance = role.auspiciousGuidance, !guidance.isEmpty else {
            return "\(role.dayName) สัมพันธ์กับ\(role.roleName)ในการวิเคราะห์ฤกษ์ของคุณ"
        }
        return guidance
    }

    private func detailSection(title: String, text: String) -> some View {
        VStack(alignment: .leading, spacing: 5) {
            Text(title)
                .font(.sarabun(13, weight: .bold))
                .foregroundStyle(Color.brandGold)
            Text(text)
                .font(.sarabun(14))
                .foregroundStyle(Color.brandCream)
                .fixedSize(horizontal: false, vertical: true)
        }
    }
}

private extension TaksaRole {
    var wheelDayName: String {
        dayName.components(separatedBy: " (").first ?? dayName
    }
}

nonisolated private struct TaksaPointerShape: Shape {
    func path(in rect: CGRect) -> Path {
        var path = Path()
        path.move(to: CGPoint(x: rect.midX, y: rect.maxY))
        path.addLine(to: CGPoint(x: rect.minX, y: rect.minY))
        path.addLine(to: CGPoint(x: rect.maxX, y: rect.minY))
        path.closeSubpath()
        return path
    }
}

private struct TaksaHeroGrid: View {
    let heroRoles: [String: TaksaRole]
    private let items: [(String, String)] = [
        ("sri", "⭐ ศรี"), ("dech", "👑 เดช"),
        ("montri", "🤝 มนตรี"), ("kalakini", "⚠ กาลกิณี")
    ]
    private let columns = [GridItem(.flexible(), spacing: 8), GridItem(.flexible(), spacing: 8)]

    var body: some View {
        LazyVGrid(columns: columns, spacing: 8) {
            ForEach(items, id: \.0) { key, label in
                if let role = heroRoles[key] {
                    VStack(alignment: .leading, spacing: 3) {
                        Text("\(label) \(role.score > 0 ? "+\(role.score)" : "\(role.score)")")
                            .font(.sarabun(11, weight: .bold))
                            .foregroundStyle(taksaRoleColor(role))
                        Text(role.dayName)
                            .font(.sarabun(11, weight: .semibold))
                            .foregroundStyle(Color.brandCream)
                    }
                    .frame(maxWidth: .infinity, minHeight: 48, alignment: .leading)
                    .padding(9)
                    .background(taksaRoleColor(role).opacity(0.10))
                    .overlay(RoundedRectangle(cornerRadius: 12).stroke(taksaRoleColor(role).opacity(0.42)))
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }
            }
        }
    }
}

private func taksaSectorColor(_ role: TaksaRole, index: Int) -> Color {
    if role.roleName.contains("ศรี") { return Color(red: 0.29, green: 0.29, blue: 0.08) }
    if role.roleName.contains("เดช") { return Color(red: 0.34, green: 0.21, blue: 0.09) }
    if role.roleName.contains("มนตรี") { return Color(hex: "#3D2517") }
    if role.roleName.contains("กาล") { return Color(red: 0.33, green: 0.10, blue: 0.10) }
    return index.isMultiple(of: 2)
        ? Color(hex: "#382014")
        : Color(hex: "#26160E")
}

private func taksaRoleColor(_ role: TaksaRole) -> Color {
    if role.roleName.contains("ศรี") { return Color(red: 0.96, green: 0.90, blue: 0.42) }
    if role.roleName.contains("เดช") { return Color(red: 1.0, green: 0.79, blue: 0.53) }
    if role.roleName.contains("มนตรี") { return Color(hex: "#FFE0A0") }
    if role.roleName.contains("กาล") { return Color(red: 1.0, green: 0.65, blue: 0.65) }
    return Color.brandGold
}

private struct ThaiTimeField: View {
    @Binding var selection: Date
    @Binding var isSet: Bool
    @State private var showsPicker = false
    @State private var shimmerPhase: CGFloat = 0

    private let magicBorderColors = [
        Color.brandGold,
        Color(hex: "#FFF0A3"),
        Color(hex: "#FFDF85"),
        Color(hex: "#A37C10"),
        Color.brandGold
    ]

    private let magicBgGradient = LinearGradient(
        colors: [
            Color(hex: "#26160E"),
            Color(hex: "#160C08"),
            Color(hex: "#26160E")
        ],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )

    var body: some View {
        Button { showsPicker = true } label: {
            HStack(spacing: 12) {
                ZStack {
                    Circle()
                        .fill(Color.brandGold.opacity(0.18))
                        .frame(width: 38, height: 38)
                    Image(systemName: "clock.fill")
                        .font(.system(size: 18))
                        .foregroundStyle(Color.brandGold)
                }
                VStack(alignment: .leading, spacing: 3) {
                    Text("✨ เวลาเกิด · กุญแจสำคัญของทักษาจร")
                        .font(.sarabun(11, weight: .bold))
                        .foregroundStyle(Color.brandGold)
                    Text(isSet ? ThaiDateCodec.shared.displayTime(selection) : "แตะเพื่อระบุเวลาเกิด")
                        .font(.sarabun(16, weight: .bold))
                        .foregroundStyle(isSet ? Color.white : Color.brandCream.opacity(0.70))
                }
                Spacer()
                if isSet {
                    Button {
                        isSet = false
                    } label: {
                        Image(systemName: "xmark.circle.fill")
                            .font(.system(size: 18))
                            .foregroundStyle(Color.brandCream.opacity(0.6))
                    }
                    .buttonStyle(.plain)
                }
            }
            .padding(.horizontal, 14)
            .frame(maxWidth: .infinity, minHeight: 72, alignment: .leading)
            .background(magicBgGradient)
            .clipShape(RoundedRectangle(cornerRadius: 13))
            .overlay(
                RoundedRectangle(cornerRadius: 13)
                    .stroke(
                        LinearGradient(
                            colors: magicBorderColors,
                            startPoint: UnitPoint(x: shimmerPhase - 1.0, y: 0),
                            endPoint: UnitPoint(x: shimmerPhase + 1.0, y: 1)
                        ),
                        lineWidth: isSet ? 1.8 : 1.2
                    )
            )
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .onAppear {
            withAnimation(.linear(duration: 3.2).repeatForever(autoreverses: false)) {
                shimmerPhase = 2.0
            }
        }
        .sheet(isPresented: $showsPicker) {
            NavigationStack {
                DatePicker("เวลาเกิด", selection: $selection, displayedComponents: .hourAndMinute)
                    .datePickerStyle(.wheel)
                    .labelsHidden()
                    .environment(\.locale, Locale(identifier: "th_TH"))
                    .padding()
                    .navigationTitle("เลือกเวลาเกิด")
                    .navigationBarTitleDisplayMode(.inline)
                    .toolbar {
                        ToolbarItem(placement: .cancellationAction) { Button("ยกเลิก") { showsPicker = false } }
                        ToolbarItem(placement: .confirmationAction) {
                            Button("ตกลง") { isSet = true; showsPicker = false }
                        }
                    }
            }
            .presentationDetents([.medium])
            .preferredColorScheme(.dark)
            .tint(Color.brandGold)
        }
    }
}

private struct ThaiDateField: View {
    let label: String
    @Binding var selection: Date
    let range: ClosedRange<Date>
    var isSet: Binding<Bool>? = nil
    var placeholder = "แตะเพื่อเลือก"
    var startWithYearSelection: Bool = false
    @State private var showsPicker = false

    var body: some View {
        Button { showsPicker = true } label: {
            HStack(spacing: 8) {
                ZStack {
                    Circle()
                        .fill(Color.brandGold.opacity(0.18))
                        .frame(width: 34, height: 34)
                    Image(systemName: "calendar")
                        .font(.system(size: 16))
                        .foregroundStyle(Color.brandGold)
                }
                VStack(alignment: .leading, spacing: 3) {
                    Text(label)
                        .font(.sarabun(11, weight: .bold))
                        .foregroundStyle(Color.brandGold)
                    Text(isSet?.wrappedValue == false ? placeholder : ThaiDateCodec.shared.displayDate(selection))
                        .font(.sarabun(14, weight: .bold))
                        .foregroundStyle(isSet?.wrappedValue == false ? Color.brandCream.opacity(0.70) : Color.white)
                        .lineLimit(1)
                        .minimumScaleFactor(0.60)
                }
                Spacer(minLength: 0)
            }
            .padding(.horizontal, 10)
            .frame(maxWidth: .infinity, minHeight: 72, alignment: .leading)
            .background(Color(hex: "#1F120B"))
            .clipShape(RoundedRectangle(cornerRadius: 13))
            .overlay(
                RoundedRectangle(cornerRadius: 13)
                    .stroke(Color.brandGold.opacity(0.38), lineWidth: 1.2)
            )
            .contentShape(Rectangle())
        }
        .frame(maxWidth: .infinity)
        .buttonStyle(.plain)
        .sheet(isPresented: $showsPicker) {
            ThaiDatePickerSheet(
                label: label,
                selection: $selection,
                range: range,
                hasInitialSelection: isSet?.wrappedValue ?? true,
                startWithYearSelection: startWithYearSelection
            ) {
                isSet?.wrappedValue = true
                showsPicker = false
            }
            .presentationDetents([.fraction(0.88), .large])
            .presentationDragIndicator(.visible)
        }
    }
}

private enum CalendarPickerStep {
    case year, month, day
}

nonisolated private struct CaretPointerShape: Shape {
    var fraction: CGFloat
    var animatableData: CGFloat {
        get { fraction }
        set { fraction = newValue }
    }
    func path(in rect: CGRect) -> Path {
        var path = Path()
        let cx = rect.width * fraction
        let w: CGFloat = 18
        let h: CGFloat = 8
        path.move(to: CGPoint(x: cx, y: 0))
        path.addLine(to: CGPoint(x: cx + w / 2, y: h))
        path.addLine(to: CGPoint(x: cx - w / 2, y: h))
        path.closeSubpath()
        return path
    }
}

private struct ThaiDatePickerSheet: View {
    let label: String
    @Binding var selection: Date
    let range: ClosedRange<Date>
    var hasInitialSelection: Bool = true
    var startWithYearSelection: Bool = false
    let onDone: () -> Void
    @Environment(\.dismiss) private var dismiss

    @State private var selectedDate: Date
    @State private var hasPickedDay: Bool
    @State private var visibleYear: Int
    @State private var visibleMonth: Int
    @State private var pickerStep: CalendarPickerStep

    private let thaiMonths = ["ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."]
    private let thaiFullMonths = ["มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน", "กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม"]
    private let weekdays = ["อา", "จ", "อ", "พ", "พฤ", "ศ", "ส"]

    private var calendar: Calendar {
        var cal = Calendar(identifier: .gregorian)
        cal.timeZone = TimeZone(identifier: "Asia/Bangkok") ?? .current
        return cal
    }

    init(
        label: String,
        selection: Binding<Date>,
        range: ClosedRange<Date>,
        hasInitialSelection: Bool = true,
        startWithYearSelection: Bool = false,
        onDone: @escaping () -> Void
    ) {
        self.label = label
        self._selection = selection
        self.range = range
        self.hasInitialSelection = hasInitialSelection
        self.startWithYearSelection = startWithYearSelection
        self.onDone = onDone

        var cal = Calendar(identifier: .gregorian)
        cal.timeZone = TimeZone(identifier: "Asia/Bangkok") ?? .current
        let initial = selection.wrappedValue
        let comps = cal.dateComponents([.year, .month, .day], from: initial)
        _selectedDate = State(initialValue: initial)
        _hasPickedDay = State(initialValue: hasInitialSelection)
        _visibleYear = State(initialValue: comps.year ?? 2026)
        _visibleMonth = State(initialValue: comps.month ?? 1)
        _pickerStep = State(initialValue: startWithYearSelection ? .year : .day)
    }

    private var minYear: Int { calendar.component(.year, from: range.lowerBound) }
    private var maxYear: Int { calendar.component(.year, from: range.upperBound) }
    private var yearsList: [Int] { Array((minYear...maxYear).reversed()) }

    private var minDateStart: Date { calendar.startOfDay(for: range.lowerBound) }
    private var maxDateStart: Date { calendar.startOfDay(for: range.upperBound) }

    private var arrowFraction: CGFloat {
        switch pickerStep {
        case .year: return 1.0 / 6.0
        case .month: return 3.0 / 6.0
        case .day: return 5.0 / 6.0
        }
    }

    var body: some View {
        NavigationStack {
            VStack(spacing: 12) {
                // Header Display
                VStack(alignment: .leading, spacing: 5) {
                    HStack {
                        Text(label)
                            .font(.sarabun(17, weight: .bold))
                            .foregroundStyle(Color.brandGold)
                        Spacer()
                        Button {
                            dismiss()
                        } label: {
                            Image(systemName: "xmark.circle.fill")
                                .font(.system(size: 20))
                                .foregroundStyle(Color.brandCream.opacity(0.45))
                        }
                    }

                    Text(hasPickedDay ? "วันที่เลือก: \(ThaiDateCodec.shared.displayDate(selectedDate))" : "เลือกปี → เดือน → วัน")
                        .font(.sarabun(hasPickedDay ? 20 : 15, weight: .bold))
                        .foregroundStyle(hasPickedDay ? Color.brandCream : Color.brandGold)
                }
                .padding(.horizontal, 16)
                .padding(.top, 16)

                Divider()
                    .overlay(Color.brandGold.opacity(0.35))
                    .padding(.horizontal, 16)

                // 3-Step Selection Tabs
                HStack(spacing: 8) {
                    stepTabButton(
                        step: .year,
                        title: "1. ปี พ.ศ.",
                        value: String(visibleYear + 543),
                        isActive: pickerStep == .year
                    )
                    stepTabButton(
                        step: .month,
                        title: "2. เดือน",
                        value: thaiMonths[max(0, min(11, visibleMonth - 1))],
                        isActive: pickerStep == .month
                    )
                    stepTabButton(
                        step: .day,
                        title: "3. วัน",
                        value: hasPickedDay && calendar.component(.year, from: selectedDate) == visibleYear && calendar.component(.month, from: selectedDate) == visibleMonth ? "วันที่ \(calendar.component(.day, from: selectedDate))" : "เลือกวัน",
                        isActive: pickerStep == .day
                    )
                }
                .padding(.horizontal, 16)

                // Speech Bubble Pointer + Active Step Card
                VStack(spacing: 0) {
                    CaretPointerShape(fraction: arrowFraction)
                        .fill(Color(hex: "#160C08"))
                        .frame(height: 8)
                        .overlay(
                            CaretPointerShape(fraction: arrowFraction)
                                .stroke(Color.brandGold.opacity(0.4), lineWidth: 1)
                        )
                        .animation(.spring(response: 0.35, dampingFraction: 0.75), value: arrowFraction)

                    VStack(spacing: 12) {
                        switch pickerStep {
                        case .year:
                            yearPickerContent
                        case .month:
                            monthPickerContent
                        case .day:
                            dayPickerContent
                        }
                    }
                    .padding(14)
                    .frame(maxWidth: .infinity)
                    .background(Color(hex: "#160C08"))
                    .clipShape(RoundedRectangle(cornerRadius: 16))
                    .overlay(
                        RoundedRectangle(cornerRadius: 16)
                            .stroke(Color.brandGold.opacity(0.4), lineWidth: 1)
                    )
                }
                .padding(.horizontal, 16)

                Spacer(minLength: 4)

                // Bottom Action Buttons
                HStack {
                    Button("ยกเลิก") {
                        dismiss()
                    }
                    .font(.sarabun(15))
                    .foregroundStyle(Color.brandCream.opacity(0.6))
                    .padding(.horizontal, 14)
                    .padding(.vertical, 8)

                    Spacer()

                    Button {
                        selection = selectedDate
                        onDone()
                        dismiss()
                    } label: {
                        Text(hasPickedDay ? "ตกลง" : "เลือกวันก่อน")
                            .font(.sarabun(15, weight: .bold))
                            .foregroundStyle(hasPickedDay ? Color.brandGreen : Color.brandCream.opacity(0.4))
                            .padding(.horizontal, 24)
                            .padding(.vertical, 10)
                            .background(hasPickedDay ? Color.brandGold : Color.brandGold.opacity(0.2))
                            .clipShape(RoundedRectangle(cornerRadius: 10))
                    }
                    .disabled(!hasPickedDay)
                }
                .padding(.horizontal, 16)
                .padding(.bottom, 14)
            }
            .background(Color.brandBackground.ignoresSafeArea())
        }
        .preferredColorScheme(.dark)
    }

    private func stepTabButton(
        step: CalendarPickerStep,
        title: String,
        value: String,
        isActive: Bool
    ) -> some View {
        Button {
            UIImpactFeedbackGenerator(style: .light).impactOccurred()
            withAnimation(.spring(response: 0.35, dampingFraction: 0.75)) {
                pickerStep = step
            }
        } label: {
            VStack(spacing: 2) {
                Text(title)
                    .font(.sarabun(10, weight: .bold))
                    .foregroundStyle(isActive ? Color.brandGreen.opacity(0.85) : Color.brandCream.opacity(0.6))
                Text(verbatim: value)
                    .font(.sarabun(13, weight: .bold))
                    .foregroundStyle(isActive ? Color.brandGreen : Color.brandCream)
                    .lineLimit(1)
            }
            .frame(maxWidth: .infinity, minHeight: 52)
            .background(isActive ? Color.brandGold : Color(hex: "#1F120B"))
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .overlay(
                RoundedRectangle(cornerRadius: 12)
                    .stroke(isActive ? Color.brandGold : Color.brandGold.opacity(0.35), lineWidth: isActive ? 1.5 : 1)
            )
        }
        .buttonStyle(.plain)
    }

    // Year Picker Content
    private var yearPickerContent: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("1  เลือกปี พ.ศ.")
                .font(.sarabun(15, weight: .bold))
                .foregroundStyle(Color.brandGold)
            Text("แตะเลือกปี พ.ศ. ด้านล่างเพื่อไปยังขั้นตอนเลือกเดือน")
                .font(.sarabun(12))
                .foregroundStyle(Color.brandCream.opacity(0.6))

            ScrollViewReader { proxy in
                ScrollView {
                    LazyVGrid(columns: Array(repeating: GridItem(.flexible(), spacing: 8), count: 3), spacing: 8) {
                        ForEach(yearsList, id: \.self) { year in
                            let isSelected = year == visibleYear
                            Button {
                                UIImpactFeedbackGenerator(style: .light).impactOccurred()
                                visibleYear = year
                                withAnimation(.spring(response: 0.35, dampingFraction: 0.75)) {
                                    pickerStep = .month
                                }
                            } label: {
                                Text(verbatim: "\(year + 543)")
                                    .font(.sarabun(14, weight: isSelected ? .bold : .medium))
                                    .foregroundStyle(isSelected ? Color.brandGold : Color.brandCream)
                                    .frame(maxWidth: .infinity, minHeight: 40)
                                    .background(isSelected ? Color.brandGold.opacity(0.22) : Color.clear)
                                    .clipShape(RoundedRectangle(cornerRadius: 10))
                                    .overlay(
                                        RoundedRectangle(cornerRadius: 10)
                                            .stroke(isSelected ? Color.brandGold : Color.brandGold.opacity(0.25), lineWidth: 1)
                                    )
                            }
                            .buttonStyle(.plain)
                            .id(year)
                        }
                    }
                    .padding(.vertical, 4)
                }
                .frame(maxHeight: 240)
                .onAppear {
                    DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
                        proxy.scrollTo(visibleYear, anchor: .center)
                    }
                }
            }

            Button {
                withAnimation(.spring(response: 0.35, dampingFraction: 0.75)) {
                    pickerStep = .month
                }
            } label: {
                Text("ไปเลือกเดือน (พ.ศ. \(String(visibleYear + 543))) →")
                    .font(.sarabun(13, weight: .semibold))
                    .foregroundStyle(Color.brandGold)
                    .frame(maxWidth: .infinity, minHeight: 38)
                    .background(Color.brandGold.opacity(0.15))
                    .clipShape(RoundedRectangle(cornerRadius: 10))
                    .overlay(
                        RoundedRectangle(cornerRadius: 10)
                            .stroke(Color.brandGold.opacity(0.5), lineWidth: 1)
                    )
            }
            .buttonStyle(.plain)
        }
    }

    // Month Picker Content
    private var monthPickerContent: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text("2  เลือกเดือน")
                    .font(.sarabun(15, weight: .bold))
                    .foregroundStyle(Color.brandGold)
                Spacer()
                Text("ปี พ.ศ. \(String(visibleYear + 543))")
                    .font(.sarabun(13))
                    .foregroundStyle(Color.brandCream.opacity(0.6))
            }
            Text("แตะเลือกเดือนด้านล่างเพื่อไปยังขั้นตอนเลือกวัน")
                .font(.sarabun(12))
                .foregroundStyle(Color.brandCream.opacity(0.6))

            LazyVGrid(columns: Array(repeating: GridItem(.flexible(), spacing: 8), count: 3), spacing: 8) {
                ForEach(1...12, id: \.self) { m in
                    let isSelected = m == visibleMonth
                    let isEnabled = isMonthInRange(m)
                    Button {
                        if isEnabled {
                            UIImpactFeedbackGenerator(style: .light).impactOccurred()
                            visibleMonth = m
                            withAnimation(.spring(response: 0.35, dampingFraction: 0.75)) {
                                pickerStep = .day
                            }
                        }
                    } label: {
                        Text(thaiMonths[m - 1])
                            .font(.sarabun(13, weight: isSelected ? .bold : .regular))
                            .foregroundStyle(
                                isEnabled
                                    ? (isSelected ? Color.brandGold : Color.brandCream)
                                    : Color.brandCream.opacity(0.25)
                            )
                            .frame(maxWidth: .infinity, minHeight: 40)
                            .background(isSelected ? Color.brandGold.opacity(0.18) : Color.clear)
                            .clipShape(RoundedRectangle(cornerRadius: 10))
                            .overlay(
                                RoundedRectangle(cornerRadius: 10)
                                    .stroke(
                                        isEnabled
                                            ? (isSelected ? Color.brandGold : Color.brandGold.opacity(0.25))
                                            : Color.clear,
                                        lineWidth: 1
                                    )
                            )
                            .contentShape(Rectangle())
                    }
                    .buttonStyle(.plain)
                    .frame(maxWidth: .infinity, minHeight: 40)
                    .disabled(!isEnabled)
                }
            }
            .padding(.vertical, 4)

            Button {
                withAnimation(.spring(response: 0.35, dampingFraction: 0.75)) {
                    pickerStep = .day
                }
            } label: {
                Text("ไปเลือกวัน (\(thaiMonths[visibleMonth - 1]) \(String(visibleYear + 543))) →")
                    .font(.sarabun(13, weight: .semibold))
                    .foregroundStyle(Color.brandGold)
                    .frame(maxWidth: .infinity, minHeight: 38)
                    .background(Color.brandGold.opacity(0.15))
                    .clipShape(RoundedRectangle(cornerRadius: 10))
                    .overlay(
                        RoundedRectangle(cornerRadius: 10)
                            .stroke(Color.brandGold.opacity(0.5), lineWidth: 1)
                    )
            }
            .buttonStyle(.plain)
        }
    }

    // Day Picker Content
    private var dayPickerContent: some View {
        VStack(spacing: 8) {
            // Month switcher
            HStack {
                Button {
                    stepMonth(by: -1)
                } label: {
                    Image(systemName: "chevron.left")
                        .font(.system(size: 14, weight: .bold))
                        .foregroundStyle(canStepMonth(by: -1) ? Color.brandGold : Color.brandCream.opacity(0.25))
                        .frame(width: 32, height: 32)
                }
                .disabled(!canStepMonth(by: -1))

                Spacer()

                VStack(spacing: 2) {
                    Text("3  เลือกวัน")
                        .font(.sarabun(14, weight: .bold))
                        .foregroundStyle(Color.brandGold)
                    Text("\(thaiFullMonths[visibleMonth - 1]) \(String(visibleYear + 543))")
                        .font(.sarabun(12))
                        .foregroundStyle(Color.brandCream.opacity(0.65))
                }

                Spacer()

                Button {
                    stepMonth(by: 1)
                } label: {
                    Image(systemName: "chevron.right")
                        .font(.system(size: 14, weight: .bold))
                        .foregroundStyle(canStepMonth(by: 1) ? Color.brandGold : Color.brandCream.opacity(0.25))
                        .frame(width: 32, height: 32)
                }
                .disabled(!canStepMonth(by: 1))
            }

            // Weekdays Row
            HStack {
                ForEach(weekdays, id: \.self) { day in
                    Text(day)
                        .font(.sarabun(11, weight: .bold))
                        .foregroundStyle(Color.brandCream.opacity(0.6))
                        .frame(maxWidth: .infinity)
                }
            }

            // 7-Column Days Grid
            let startOfMonth = calendar.date(from: DateComponents(year: visibleYear, month: visibleMonth, day: 1)) ?? Date()
            let daysInMonth = calendar.range(of: .day, in: .month, for: startOfMonth)?.count ?? 30
            let firstWeekday = calendar.component(.weekday, from: startOfMonth) // 1 = Sun, 2 = Mon...
            let firstOffset = (firstWeekday - 1) % 7

            LazyVGrid(columns: Array(repeating: GridItem(.flexible(), spacing: 4), count: 7), spacing: 4) {
                ForEach(0..<42, id: \.self) { cell in
                    let dayNumber = cell - firstOffset + 1
                    if dayNumber >= 1 && dayNumber <= daysInMonth {
                        let cellDate = calendar.date(from: DateComponents(year: visibleYear, month: visibleMonth, day: dayNumber)) ?? Date()
                        let isEnabled = cellDate >= minDateStart && cellDate <= maxDateStart
                        let isSelected = hasPickedDay && calendar.isDate(cellDate, inSameDayAs: selectedDate)

                        Button {
                            if isEnabled {
                                UIImpactFeedbackGenerator(style: .light).impactOccurred()
                                selectedDate = cellDate
                                hasPickedDay = true
                            }
                        } label: {
                            Text("\(dayNumber)")
                                .font(.sarabun(13, weight: isSelected ? .bold : .regular))
                                .foregroundStyle(
                                    isSelected
                                        ? Color.brandGreen
                                        : (isEnabled ? Color.brandCream : Color.brandCream.opacity(0.25))
                                )
                                .frame(maxWidth: .infinity, minHeight: 32)
                                .background(isSelected ? Color.brandGold : Color.clear)
                                .clipShape(Circle())
                        }
                        .buttonStyle(.plain)
                        .disabled(!isEnabled)
                    } else {
                        Color.clear.frame(minHeight: 32)
                    }
                }
            }
        }
    }

    private func isMonthInRange(_ m: Int) -> Bool {
        guard let start = calendar.date(from: DateComponents(year: visibleYear, month: m, day: 1)),
              let days = calendar.range(of: .day, in: .month, for: start)?.count,
              let end = calendar.date(from: DateComponents(year: visibleYear, month: m, day: days)) else {
            return false
        }
        return !(end < minDateStart || start > maxDateStart)
    }

    private func canStepMonth(by delta: Int) -> Bool {
        var comps = DateComponents()
        comps.month = delta
        guard let cur = calendar.date(from: DateComponents(year: visibleYear, month: visibleMonth, day: 1)),
              let next = calendar.date(byAdding: comps, to: cur) else { return false }
        let nextYear = calendar.component(.year, from: next)
        let nextMonth = calendar.component(.month, from: next)
        guard let start = calendar.date(from: DateComponents(year: nextYear, month: nextMonth, day: 1)),
              let days = calendar.range(of: .day, in: .month, for: start)?.count,
              let end = calendar.date(from: DateComponents(year: nextYear, month: nextMonth, day: days)) else { return false }
        return !(end < minDateStart || start > maxDateStart)
    }

    private func stepMonth(by delta: Int) {
        var comps = DateComponents()
        comps.month = delta
        guard let cur = calendar.date(from: DateComponents(year: visibleYear, month: visibleMonth, day: 1)),
              let next = calendar.date(byAdding: comps, to: cur) else { return }
        visibleYear = calendar.component(.year, from: next)
        visibleMonth = calendar.component(.month, from: next)
    }
}

private struct AuspiciousResultsView: View {
    let result: AuspiciousSearchResponse
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            ZStack {
                Color.brandBackground.ignoresSafeArea()
                ScrollView {
                    LazyVStack(spacing: 14) {
                        VStack(alignment: .leading, spacing: 5) {
                            Text("\(result.recommendations.count) วันที่ดีที่สุด")
                                .font(.sarabun(22, weight: .bold)).foregroundStyle(Color.brandGold)
                            Text("สำหรับ \(result.activity.nameTH)").font(.sarabun(15, weight: .semibold))
                            Text("ตรวจ \(result.evaluatedDays) วัน · ตัดวันที่ไม่ผ่าน \(result.excludedDays) วัน")
                                .font(.sarabun(12)).foregroundStyle(Color.brandCream.opacity(0.65))
                        }
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .luxuryCard()

                        if result.recommendations.isEmpty {
                            Text("ไม่พบวันที่ผ่านทุกเงื่อนไขในช่วงนี้ ลองเปลี่ยนช่วงวันที่หรือตัวเลือก")
                                .frame(maxWidth: .infinity, alignment: .leading)
                                .luxuryCard()
                        } else {
                            ForEach(Array(result.recommendations.enumerated()), id: \.element.id) { index, day in
                                RecommendationView(day: day, rank: index + 1)
                            }
                        }

                        VStack(alignment: .leading, spacing: 6) {
                            Text("หลักที่ใช้").font(.sarabun(16, weight: .bold)).foregroundStyle(Color.brandGold)
                            Text(result.ruleSets.isEmpty ? "กฎฤกษ์ตามกิจกรรมที่เลือก" : result.ruleSets.joined(separator: ", "))
                            if !result.activity.ruleScopeNote.isEmpty {
                                Text(result.activity.ruleScopeNote).font(.sarabun(12)).foregroundStyle(Color.brandCream.opacity(0.7))
                            }
                            if let safety = result.activity.safetyNote, !safety.isEmpty {
                                Text("⚠ \(safety)").font(.sarabun(12)).foregroundStyle(Color.orange)
                            }
                        }
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .luxuryCard()
                    }
                    .padding(16)
                    .padding(.bottom, 24)
                }
            }
            .navigationTitle("ผลการดูฤกษ์")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) { Button("เสร็จสิ้น") { dismiss() } }
            }
            .font(.sarabun(14))
            .tint(.brandGold)
            .preferredColorScheme(.dark)
        }
    }
}

private struct RecommendationView: View {
    let day: AuspiciousDateResult
    let rank: Int
    @State private var expanded = false

    var body: some View {
        VStack(alignment: .leading, spacing: 11) {
            HStack(alignment: .top, spacing: 12) {
                Text("\(rank)")
                    .font(.sarabun(18, weight: .bold)).foregroundStyle(Color.brandGreen)
                    .frame(width: 42, height: 42).background(Color.brandGold).clipShape(Circle())
                VStack(alignment: .leading, spacing: 3) {
                    Text(rank == 1 ? "แนะนำที่สุด" : rank == 2 ? "เหมาะมาก" : "วันที่แนะนำ")
                        .font(.sarabun(11, weight: .bold)).foregroundStyle(Color.brandGold)
                    Text(day.weekday).font(.sarabun(19, weight: .bold)).foregroundStyle(Color.brandCream)
                    Text(day.thaiDate).foregroundStyle(Color.brandCream)
                    Text(day.lunarLabel + (day.isWanPhra ? " · วันพระ" : ""))
                        .font(.sarabun(12)).foregroundStyle(Color.brandCream.opacity(0.65))
                    if let important = day.importantDay, !important.isEmpty {
                        Text("✦ \(important)").font(.sarabun(12, weight: .semibold)).foregroundStyle(Color.brandGold)
                    }
                }
                Spacer()
                VStack(spacing: 0) {
                    Text("\(day.score)").font(.sarabun(27, weight: .bold)).foregroundStyle(Color.brandGold)
                    Text("คะแนน").font(.sarabun(10)).foregroundStyle(Color.brandCream.opacity(0.6))
                }
            }
            Divider().overlay(Color.brandGold.opacity(0.18))
            Text(day.rating).font(.sarabun(14, weight: .bold)).foregroundStyle(Color.brandCream)
            ForEach(Array(day.reasons.prefix(expanded ? day.reasons.count : 3).enumerated()), id: \.offset) { _, reason in
                VStack(alignment: .leading, spacing: 2) {
                    Text("\(reason.impact == "caution" ? "⚠" : "✓") \(reason.name)")
                        .font(.sarabun(13, weight: .semibold))
                        .foregroundStyle(reason.impact == "caution" ? Color.orange : Color(hex: "#FFDF85"))
                    if expanded {
                        Text(reason.meaning.isEmpty ? reason.explanation : reason.meaning)
                            .font(.sarabun(12)).foregroundStyle(Color.brandCream.opacity(0.68))
                    }
                }
            }
            if !day.reasons.isEmpty {
                Button(expanded ? "ซ่อนความหมาย" : "ดูความหมายของฤกษ์") { withAnimation { expanded.toggle() } }
                    .font(.sarabun(12, weight: .semibold))
                    .frame(maxWidth: .infinity, alignment: .trailing)
            }

            if let advice = day.outfitAdvice, !advice.auspicious.isEmpty {
                Divider().overlay(Color.brandGold.opacity(0.18))
                OutfitColorSectionView(
                    title: "สีที่เสริม",
                    sets: advice.auspicious,
                    positive: true
                )
                if !advice.inauspicious.isEmpty {
                    OutfitColorSectionView(
                        title: "สีที่ควรเลี่ยง",
                        sets: advice.inauspicious,
                        positive: false
                    )
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .luxuryCard()
        .overlay(RoundedRectangle(cornerRadius: 22).stroke(Color.brandGold.opacity(rank == 1 ? 0.85 : 0.25), lineWidth: rank == 1 ? 2 : 1))
    }
}

private struct OutfitColorSectionView: View {
    let title: String
    let sets: [OutfitColorSet]
    let positive: Bool

    var body: some View {
        if !sets.isEmpty {
            VStack(alignment: .leading, spacing: 6) {
                Text(title)
                    .font(.sarabun(12, weight: .bold))
                    .foregroundStyle(positive ? Color.brandCream : Color.brandCream.opacity(0.7))

                HStack(alignment: .top) {
                    ForEach(sets) { set in
                        let displayLabel = labelForContext(set.context, labelTH: set.labelTH)
                        Spacer(minLength: 0)
                        OutfitColorGroupView(
                            label: displayLabel,
                            colors: set.colors,
                            positive: positive
                        )
                        Spacer(minLength: 0)
                    }
                }
                .frame(maxWidth: .infinity)
            }
        }
    }

    private func labelForContext(_ context: String, labelTH: String) -> String {
        switch context {
        case "appointment":
            return positive ? "เสริมจากวันนี้" : "เลี่ยงจากวันนี้"
        case "birth_day":
            return positive ? "เสริมจากวันเกิด" : "เลี่ยงจากวันเกิด"
        case "progressed_age":
            return positive ? "เสริมจากอายุย่าง" : "เลี่ยงจากอายุย่าง"
        default:
            return labelTH
        }
    }
}

private struct OutfitColorGroupView: View {
    let label: String
    let colors: [OutfitColor]
    let positive: Bool

    private var bracketWidth: CGFloat {
        if colors.count >= 3 { return 65 }
        if colors.count == 2 { return 42 }
        return 20
    }

    var body: some View {
        VStack(alignment: .center, spacing: 3) {
            HStack(spacing: 4) {
                ForEach(colors) { color in
                    Circle()
                        .fill(Color(hex: color.hex))
                        .frame(width: 19, height: 19)
                        .overlay(Circle().stroke(Color.white.opacity(0.45), lineWidth: 1))
                }
            }

            DownwardBracketShape()
                .stroke(
                    positive ? Color.brandGold.opacity(0.6) : Color(red: 1, green: 0.55, blue: 0.55).opacity(0.6),
                    style: StrokeStyle(lineWidth: 1, lineCap: .round, lineJoin: .round)
                )
                .frame(width: bracketWidth, height: 7)

            Text(label)
                .font(.sarabun(11, weight: .medium))
                .foregroundStyle(positive ? Color.brandCream : Color(red: 1, green: 0.7, blue: 0.7))
                .lineLimit(1)
        }
    }
}

nonisolated private struct DownwardBracketShape: Shape {
    func path(in rect: CGRect) -> Path {
        var path = Path()
        let w = rect.width
        let h = rect.height
        let midX = rect.midX
        let lineY = h - 3.5
        let leftX: CGFloat = 2
        let rightX = w - 2

        path.move(to: CGPoint(x: leftX, y: 0))
        path.addLine(to: CGPoint(x: leftX, y: lineY))
        path.addLine(to: CGPoint(x: midX - 3.5, y: lineY))
        path.addLine(to: CGPoint(x: midX, y: h))
        path.addLine(to: CGPoint(x: midX + 3.5, y: lineY))
        path.addLine(to: CGPoint(x: rightX, y: lineY))
        path.addLine(to: CGPoint(x: rightX, y: 0))

        return path
    }
}

private final class ThaiDateCodec: @unchecked Sendable {
    static let shared = ThaiDateCodec()
    private let calendar: Calendar
    private let apiDateFormatter: DateFormatter
    private let apiTimeFormatter: DateFormatter
    private let displayFormatter: DateFormatter
    private let displayTimeFormatter: DateFormatter

    private init() {
        let timeZone = TimeZone(identifier: "Asia/Bangkok") ?? .current
        var gregorian = Calendar(identifier: .gregorian)
        gregorian.timeZone = timeZone
        calendar = gregorian

        apiDateFormatter = DateFormatter()
        apiDateFormatter.calendar = gregorian
        apiDateFormatter.locale = Locale(identifier: "en_US_POSIX")
        apiDateFormatter.timeZone = timeZone
        apiDateFormatter.dateFormat = "yyyy-MM-dd"

        apiTimeFormatter = DateFormatter()
        apiTimeFormatter.calendar = gregorian
        apiTimeFormatter.locale = Locale(identifier: "en_US_POSIX")
        apiTimeFormatter.timeZone = timeZone
        apiTimeFormatter.dateFormat = "HH:mm"

        var buddhist = Calendar(identifier: .buddhist)
        buddhist.locale = Locale(identifier: "th_TH")
        buddhist.timeZone = timeZone
        displayFormatter = DateFormatter()
        displayFormatter.calendar = buddhist
        displayFormatter.locale = Locale(identifier: "th_TH")
        displayFormatter.timeZone = timeZone
        displayFormatter.dateFormat = "d MMM yyyy"

        displayTimeFormatter = DateFormatter()
        displayTimeFormatter.calendar = buddhist
        displayTimeFormatter.locale = Locale(identifier: "th_TH")
        displayTimeFormatter.timeZone = timeZone
        displayTimeFormatter.dateFormat = "HH:mm น."
    }

    var today: Date { calendar.startOfDay(for: Date()) }
    func apiDate(_ date: Date) -> String { apiDateFormatter.string(from: date) }
    func apiTime(_ date: Date) -> String { apiTimeFormatter.string(from: date) }
    func displayDate(_ date: Date) -> String { displayFormatter.string(from: date) }
    func displayTime(_ date: Date) -> String { displayTimeFormatter.string(from: date) }
    func addDays(_ days: Int, to date: Date) -> Date {
        calendar.date(byAdding: .day, value: days, to: calendar.startOfDay(for: date)) ?? calendar.startOfDay(for: date)
    }
    func days(from start: Date, to end: Date) -> Int { calendar.dateComponents([.day], from: start, to: end).day ?? 0 }
    func date(year: Int, month: Int, day: Int) -> Date { calendar.date(from: DateComponents(year: year, month: month, day: day)) ?? Date() }
}
