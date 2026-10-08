import SwiftUI

struct AboutTayaSheet: View {
    @Environment(\.dismiss) private var dismiss
    var articles: [ArticleItem] = []

    private let websiteLandingURL = URL(string: "https://xn--b3cu8e7ah6h.com/")!
    private let tayaPhotoURL = URL(string: "https://xn--b3cu8e7ah6h.com/images/taya_photo.jpg?v=5")

    var body: some View {
        NavigationStack {
            ZStack {
                Color.brandBackground.ignoresSafeArea()
                ScrollView {
                    VStack(alignment: .leading, spacing: 20) {
                        // Header Badge & Title
                        VStack(alignment: .center, spacing: 6) {
                            Text("✨ เกี่ยวกับเรา")
                                .font(.sarabun(12, weight: .bold))
                                .foregroundStyle(Color.brandGold)
                                .padding(.horizontal, 12)
                                .padding(.vertical, 4)
                                .background(Color.brandGold.opacity(0.15))
                                .clipShape(Capsule())
                                .overlay(Capsule().stroke(Color.brandGold.opacity(0.3), lineWidth: 1))

                            Text("แนะนำ อ.ทญา ว่องไว")
                                .font(.sarabun(22, weight: .bold))
                                .foregroundStyle(Color.brandGold)

                            Text("ผู้พัฒนาระบบชื่อดี 4 มิติ — ความเป็นมาและแรงบันดาลใจ")
                                .font(.sarabun(13))
                                .foregroundStyle(Color.brandCream.opacity(0.85))
                                .multilineTextAlignment(.center)
                        }
                        .frame(maxWidth: .infinity)
                        .padding(.top, 4)

                        // Photo Card
                        VStack(spacing: 8) {
                            AsyncImage(url: tayaPhotoURL) { phase in
                                switch phase {
                                case .success(let image):
                                    image
                                        .resizable()
                                        .aspectRatio(contentMode: .fill)
                                case .failure(_), .empty:
                                    Image(systemName: "person.crop.rectangle.fill")
                                        .resizable()
                                        .aspectRatio(contentMode: .fit)
                                        .foregroundStyle(Color.brandGold.opacity(0.4))
                                        .padding(24)
                                @unknown default:
                                    EmptyView()
                                }
                            }
                            .frame(width: 160, height: 185)
                            .clipShape(RoundedRectangle(cornerRadius: 18))
                            .overlay(RoundedRectangle(cornerRadius: 18).stroke(Color.brandGold.opacity(0.6), lineWidth: 1.5))
                            .shadow(color: Color.black.opacity(0.4), radius: 10, x: 0, y: 5)

                            Text("อ.ทญา ว่องไว ในวัยประถมศึกษา")
                                .font(.sarabun(12, weight: .semibold))
                                .foregroundStyle(Color.brandGold)

                            Text("จุดเริ่มต้นและแรงบันดาลใจในการศึกษาค้นคว้าศาสตร์แห่งชื่อมงคล")
                                .font(.sarabun(11))
                                .foregroundStyle(Color.brandCream.opacity(0.7))
                        }
                        .frame(maxWidth: .infinity)

                        Divider().overlay(Color.brandGold.opacity(0.3))

                        // Section 1: Background & Inspiration
                        VStack(alignment: .leading, spacing: 10) {
                            Text("ความเป็นมาและความสำคัญของปัญหา")
                                .font(.sarabun(16, weight: .bold))
                                .foregroundStyle(Color.brandGold)

                            Text("สมัยเรียนชั้นประถมศึกษา ครูประจำชั้นจะขานชื่อนักเรียนทุกเช้า เพื่อนบางคนมีชื่อและบุคลิกแปลกชวนขบขัน เช่น ครูเรียกชื่อ “เด็กชายมังกือ” ก็จะมีเสียงหัวเราะครึกครื้นในห้อง")
                                .font(.sarabun(13))
                                .foregroundStyle(Color.brandCream.opacity(0.9))
                                .lineSpacing(5)

                            Text("มีบางครั้งที่ชื่อของเพื่อนเลขที่ติดกัน เมื่อออกเสียงต่อกันแล้วฟังคล้ายบทสวดมนต์ จนคุณครูแซวว่า:")
                                .font(.sarabun(13))
                                .foregroundStyle(Color.brandCream.opacity(0.9))
                                .lineSpacing(5)

                            // Quote Box
                            HStack(spacing: 0) {
                                Rectangle()
                                    .fill(Color.brandGold)
                                    .frame(width: 4)
                                Text("“ชื่อแบบนี้เอามาเขียนต่อกันแล้วสวดเป่าแก้ตาแดงได้”")
                                    .font(.sarabun(14, weight: .semibold))
                                    .foregroundStyle(Color.brandGold)
                                    .padding(.horizontal, 14)
                                    .padding(.vertical, 10)
                            }
                            .background(Color.brandGold.opacity(0.08))
                            .clipShape(RoundedRectangle(cornerRadius: 8))

                            Text("คำพูดเปรียบเปรยของคุณครูในวันนั้น ที่เชื่อมโยงบทสวดมนต์ ภูมิปัญญาโบราณ และการรักษาโรค เข้ากับ 'ชื่อ' ได้กลายเป็นจุดประกายและแรงบันดาลใจให้ค้นคว้าอย่างลึกซึ้ง: เราจะตั้งชื่ออย่างไรให้ถูกต้องตามตำราและส่งเสริมชีวิตได้จริง?")
                                .font(.sarabun(13))
                                .foregroundStyle(Color.brandCream.opacity(0.9))
                                .lineSpacing(5)
                        }

                        // Section 2: Author & Data Science
                        VStack(alignment: .leading, spacing: 10) {
                            Text("เกี่ยวกับผู้เขียนและผู้พัฒนาระบบ")
                                .font(.sarabun(16, weight: .bold))
                                .foregroundStyle(Color.brandGold)

                            Text("แอป “ชื่อดี” พัฒนาขึ้นเมื่อองค์ความรู้ด้านการออกแบบชื่อตกผลึก และความสามารถของเทคโนโลยี AI มีความพร้อม โดย ทญา ว่องไว นำหลัก Data Science มาจัดโครงสร้างการสืบค้นและเปรียบเทียบชื่อ ผสานกับ 4 มิติสำคัญ:")
                                .font(.sarabun(13))
                                .foregroundStyle(Color.brandCream.opacity(0.9))
                                .lineSpacing(5)

                            VStack(alignment: .leading, spacing: 8) {
                                dimensionBullet("⭐ เลขศาสตร์ & พลังเงา", "คำนวณคู่เลขมงคลและผลรวมตามตำราชั้นสูงอย่างแม่นยำ")
                                dimensionBullet("🎯 ความหมายมงคล", "วิเคราะห์ความหมายลึกซึ้ง ตรงจุดประสงค์ เสริมมงคลชีวิต")
                                dimensionBullet("🎶 ความไพเราะของเสียง", "ลื่นไหล จังหวะสัมผัส และพลังเสียงที่เป็นสิริมงคล")
                                dimensionBullet("🔍 เอกลักษณ์เฉพาะตัว", "คงรากศัพท์และกลิ่นอายเสียงเดิมของชื่อต้นแบบที่คุณชื่นชอบ")
                            }
                            .padding(12)
                            .background(Color.brandGreen.opacity(0.4))
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                            .overlay(RoundedRectangle(cornerRadius: 12).stroke(Color.brandGold.opacity(0.3), lineWidth: 1))

                            Text("“การเลือกใช้ชื่อที่ดี จะแสดงรสนิยม ส่งเสริมภาพลักษณ์ และมอบประสบการณ์ที่ดีแก่ชีวิต”")
                                .font(.sarabun(13, weight: .medium))
                                .foregroundStyle(Color.brandCream.opacity(0.85))
                                .italic()
                                .padding(.top, 2)

                            Text("— ทญา ว่องไว")
                                .font(.sarabun(13, weight: .bold))
                                .foregroundStyle(Color.brandGold)
                                .frame(maxWidth: .infinity, alignment: .trailing)
                        }

                        Divider().overlay(Color.brandGold.opacity(0.3))

                        // CTA Button to website landing page
                        Link(destination: websiteLandingURL) {
                            HStack(spacing: 8) {
                                Image(systemName: "safari")
                                    .font(.system(size: 16, weight: .semibold))
                                Text("เข้าสู่หน้าเว็บไซต์หลัก (ชื่อดี.com)")
                                    .font(.sarabun(14, weight: .bold))
                            }
                            .foregroundStyle(Color.brandGreen)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 14)
                            .background(Color.brandGold)
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                        }
                    }
                    .padding(20)
                }
            }
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button {
                        dismiss()
                    } label: {
                        Image(systemName: "xmark.circle.fill")
                            .font(.system(size: 22))
                            .foregroundStyle(Color.brandGold)
                    }
                }
            }
            .preferredColorScheme(.dark)
        }
        .presentationDetents([.fraction(0.9), .large])
        .presentationDragIndicator(.visible)
    }

    private func dimensionBullet(_ title: String, _ desc: String) -> some View {
        HStack(alignment: .top, spacing: 6) {
            Text("•")
                .foregroundStyle(Color.brandGold)
            VStack(alignment: .leading, spacing: 1) {
                Text(title)
                    .font(.sarabun(12, weight: .bold))
                    .foregroundStyle(Color.brandGold)
                Text(desc)
                    .font(.sarabun(11))
                    .foregroundStyle(Color.brandCream.opacity(0.8))
            }
        }
    }
}

typealias MethodologyGuideSheet = AboutTayaSheet

