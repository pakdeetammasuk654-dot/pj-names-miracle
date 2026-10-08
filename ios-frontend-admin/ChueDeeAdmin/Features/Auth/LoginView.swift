import SwiftUI

struct LoginView: View {
    @ObservedObject var viewModel: AuthViewModel = .shared
    @FocusState private var focusedField: Field?

    enum Field {
        case username, password
    }

    var body: some View {
        ZStack {
            Color.brandBackground.ignoresSafeArea()

            ScrollView {
                VStack(spacing: 24) {
                    Spacer(minLength: 40)

                    // Logo & App Name
                    VStack(spacing: 12) {
                        Image("AppLogo")
                            .resizable()
                            .scaledToFit()
                            .frame(width: 90, height: 90)
                            .clipShape(RoundedRectangle(cornerRadius: 18))
                            .shadow(color: Color.brandGold.opacity(0.35), radius: 12)

                        Text("ชื่อดี Admin")
                            .font(.sarabun(26, weight: .bold))
                            .foregroundStyle(Color.brandGoldLight)

                        Text("ระบบจัดการแชตลูกค้าสำหรับ อ.ทญา ว่องไว")
                            .font(.sarabun(14))
                            .foregroundStyle(Color.brandCream.opacity(0.8))
                    }
                    .padding(.bottom, 12)

                    // Form Card
                    VStack(spacing: 16) {
                        VStack(alignment: .leading, spacing: 6) {
                            Text("ชื่อผู้ใช้")
                                .font(.sarabun(14, weight: .semibold))
                                .foregroundStyle(Color.brandCream)

                            TextField("Username", text: $viewModel.usernameInput)
                                .font(.sarabun(16))
                                .autocorrectionDisabled()
                                .textInputAutocapitalization(.never)
                                .focused($focusedField, equals: .username)
                                .padding(12)
                                .background(Color(red: 0.03, green: 0.12, blue: 0.08))
                                .clipShape(RoundedRectangle(cornerRadius: 10))
                                .overlay(RoundedRectangle(cornerRadius: 10).stroke(Color.brandGold.opacity(0.4), lineWidth: 1))
                                .foregroundStyle(Color.white)
                        }

                        VStack(alignment: .leading, spacing: 6) {
                            Text("รหัสผ่าน")
                                .font(.sarabun(14, weight: .semibold))
                                .foregroundStyle(Color.brandCream)

                            SecureField("Password", text: $viewModel.passwordInput)
                                .font(.sarabun(16))
                                .focused($focusedField, equals: .password)
                                .padding(12)
                                .background(Color(red: 0.03, green: 0.12, blue: 0.08))
                                .clipShape(RoundedRectangle(cornerRadius: 10))
                                .overlay(RoundedRectangle(cornerRadius: 10).stroke(Color.brandGold.opacity(0.4), lineWidth: 1))
                                .foregroundStyle(Color.white)
                        }

                        if let err = viewModel.errorMessage {
                            Text(err)
                                .font(.sarabun(13))
                                .foregroundStyle(Color.brandRed)
                                .frame(maxWidth: .infinity, alignment: .leading)
                        }

                        Button(action: {
                            focusedField = nil
                            Task {
                                await viewModel.login()
                            }
                        }) {
                            HStack {
                                if viewModel.isLoading {
                                    ProgressView().tint(.black).padding(.trailing, 4)
                                    Text("กำลังเข้าสู่ระบบ...")
                                } else {
                                    Text("เข้าสู่ระบบแอดมิน →")
                                }
                            }
                            .font(.sarabun(16, weight: .bold))
                            .foregroundStyle(Color.black)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 14)
                            .background(
                                LinearGradient(
                                    colors: [Color.brandGoldLight, Color.brandGold],
                                    startPoint: .topLeading,
                                    endPoint: .bottomTrailing
                                )
                            )
                            .clipShape(RoundedRectangle(cornerRadius: 10))
                            .shadow(color: Color.brandGold.opacity(0.3), radius: 6, x: 0, y: 3)
                        }
                        .disabled(viewModel.isLoading)
                        .padding(.top, 8)
                    }
                    .padding(24)
                    .background(Color.brandCard)
                    .clipShape(RoundedRectangle(cornerRadius: 18))
                    .overlay(RoundedRectangle(cornerRadius: 18).stroke(Color.brandGold.opacity(0.3), lineWidth: 1))
                    .padding(.horizontal, 20)

                    Spacer(minLength: 40)
                }
            }
        }
    }
}
