import SwiftUI
import PhotosUI

@MainActor
struct SupportChatView: View {
    @ObservedObject var viewModel: SupportChatViewModel = .shared
    @Environment(\.dismiss) private var dismiss

    @State private var customerNameInput = ""
    @State private var messageText = ""
    @State private var selectedPhotoItem: PhotosPickerItem?
    @State private var isFullScreenImagePresented = false
    @State private var fullScreenImageURL: URL?
    @State private var showDeleteConfirmation = false

    var body: some View {
        NavigationStack {
            ZStack {
                Color.brandBackground.ignoresSafeArea()

                if !viewModel.hasConversation {
                    startChatView
                } else {
                    activeChatView
                }
            }
            .navigationTitle("ปรึกษา อ.ทญา ว่องไว")
            .navigationBarTitleDisplayMode(.inline)
            .toolbarBackground(Color.brandGreen.opacity(0.9), for: .navigationBar)
            .toolbarBackground(.visible, for: .navigationBar)
            .toolbarColorScheme(.dark, for: .navigationBar)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(action: {
                        viewModel.onDisappear()
                        dismiss()
                    }) {
                        Image(systemName: "chevron.left")
                            .font(.system(size: 16, weight: .semibold))
                            .foregroundStyle(Color.brandCream)
                    }
                }

                if viewModel.hasConversation {
                    ToolbarItem(placement: .primaryAction) {
                        Menu {
                            Button("เริ่มใหม่", action: viewModel.startNewConversation)
                            Button("ลบแชตและข้อมูลทั้งหมด", role: .destructive) {
                                showDeleteConfirmation = true
                            }
                            .disabled(viewModel.isDeleting)
                        } label: {
                            Image(systemName: "ellipsis.circle")
                                .foregroundStyle(Color.brandGold)
                        }
                    }
                }
            }
            .confirmationDialog(
                "ลบแชตและข้อมูลทั้งหมด?",
                isPresented: $showDeleteConfirmation,
                titleVisibility: .visible
            ) {
                Button("ลบแชต", role: .destructive) {
                    viewModel.deleteConversation()
                }
                Button("ยกเลิก", role: .cancel) {}
            } message: {
                Text("ข้อความ รูปที่แนบ และข้อมูลแจ้งเตือนของบทสนทนานี้จะถูกลบจากเซิร์ฟเวอร์ การลบย้อนกลับไม่ได้")
            }
            .onAppear {
                viewModel.onAppear()
            }
            .onDisappear {
                viewModel.onDisappear()
            }
            .sheet(isPresented: $isFullScreenImagePresented) {
                if let url = fullScreenImageURL {
                    FullScreenImageView(url: url)
                }
            }
        }
    }

    // MARK: - Start Chat View
    private var startChatView: some View {
        ScrollView {
            VStack(spacing: 20) {
                Spacer(minLength: 10)

                Image(systemName: "sparkles")
                    .font(.system(size: 44))
                    .foregroundStyle(Color.brandGold)
                    .padding(.top, 10)

                Text("ปรึกษาเรื่องชื่อกับคุณทญา")
                    .font(.sarabun(24, weight: .bold))
                    .foregroundStyle(Color.brandCream)

                Text("ส่งคำถามและรูปประกอบ ปรึกษากับ อ.ทญา ว่องไว (Data Scientist & ผู้เชี่ยวชาญด้านศาสตร์การตั้งชื่อมงคล) ได้โดยตรง")
                    .font(.sarabun(15))
                    .foregroundStyle(Color.brandCream.opacity(0.85))
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, 20)

                VStack(alignment: .leading, spacing: 10) {
                    Text("เลือก avatar น่ารักของคุณ")
                        .font(.sarabun(15, weight: .bold))
                        .foregroundStyle(Color.brandCream)

                    HStack {
                        ForEach(ChatAvatarOption.all) { option in
                            avatarOptionButton(
                                avatarKey: option.id,
                                imagePath: option.imagePath
                            )
                            if option.id != ChatAvatarOption.all.last?.id {
                                Spacer()
                            }
                        }
                    }
                    .padding(.vertical, 4)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(.horizontal, 24)

                VStack(alignment: .leading, spacing: 8) {
                    Text("ชื่อที่ให้เราเรียก")
                        .font(.sarabun(14, weight: .semibold))
                        .foregroundStyle(Color.brandCream)

                    TextField("ชื่อเล่น หรือชื่อที่ใช้เรียก (ไม่บังคับ)", text: $customerNameInput)
                        .font(.sarabun(16))
                        .padding(14)
                        .background(Color.brandGreen.opacity(0.35))
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                        .overlay(RoundedRectangle(cornerRadius: 12).stroke(Color.brandGold.opacity(0.4), lineWidth: 1))
                        .foregroundStyle(Color.white)
                }
                .padding(.horizontal, 24)

                if let err = viewModel.errorMessage {
                    Text(err)
                        .font(.sarabun(14))
                        .foregroundStyle(Color.red.opacity(0.9))
                        .padding(.horizontal, 24)
                }

                Button(action: handleStartChat) {
                    HStack {
                        if viewModel.isLoading {
                            ProgressView()
                                .tint(Color.black)
                                .padding(.trailing, 4)
                            Text("กำลังเริ่มแชต...")
                        } else {
                            Text("เริ่มแชตกับคุณทญา →")
                        }
                    }
                    .font(.sarabun(16, weight: .bold))
                    .foregroundStyle(Color.black)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 14)
                    .background(
                        LinearGradient(
                            colors: [Color(red: 0.97, green: 0.91, blue: 0.66), Color.brandGold],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        )
                    )
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                    .shadow(color: Color.brandGold.opacity(0.3), radius: 8, x: 0, y: 4)
                }
                .disabled(viewModel.isLoading)
                .padding(.horizontal, 24)
                .padding(.top, 10)

                Spacer(minLength: 30)
            }
        }
    }

    private func handleStartChat() {
        let name = customerNameInput
        Task { @MainActor [viewModel] in
            await viewModel.startConversation(name: name)
        }
    }

    private func handleSendMessage() {
        let textToSend = messageText
        messageText = ""
        selectedPhotoItem = nil
        Task { @MainActor [viewModel] in
            await viewModel.sendCurrentMessage(body: textToSend)
        }
    }

    private func handleClearAttachedImage() {
        selectedPhotoItem = nil
        viewModel.clearAttachedImage()
    }

    private func avatarOptionButton(avatarKey: String, imagePath: String) -> some View {
        let isSelected = viewModel.customerAvatar == avatarKey
        let avatarURL = APIClient.shared.avatarURL(for: imagePath)

        return Button(action: {
            viewModel.selectAvatar(avatarKey)
        }) {
            RemoteAvatarImage(url: avatarURL) {
                Circle().fill(Color.brandGreen.opacity(0.6))
            }
            .frame(width: 52, height: 52)
            .clipShape(Circle())
            .overlay(
                Circle()
                    .stroke(isSelected ? Color.brandGold : Color.brandGold.opacity(0.25), lineWidth: isSelected ? 3 : 1.2)
            )
            .shadow(color: isSelected ? Color.brandGold.opacity(0.5) : Color.black.opacity(0.25), radius: isSelected ? 6 : 2, x: 0, y: 2)
            .scaleEffect(isSelected ? 1.08 : 1.0)
            .animation(.spring(response: 0.25, dampingFraction: 0.7), value: isSelected)
        }
        .buttonStyle(.plain)
    }

    // MARK: - Active Chat View
    private var activeChatView: some View {
        VStack(spacing: 0) {
            // Chat message list
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(spacing: 14) {
                        ForEach(viewModel.messages) { msg in
                            ChatBubbleRow(
                                message: msg,
                                customerAvatarPath: viewModel.customerAvatarPath,
                                onImageTap: { url in
                                    fullScreenImageURL = url
                                    isFullScreenImagePresented = true
                                }
                            )
                            .id(msg.id)
                        }
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 14)
                }
                .onChange(of: viewModel.messages.count) {
                    if let lastID = viewModel.messages.last?.id {
                        withAnimation(.easeOut(duration: 0.25)) {
                            proxy.scrollTo(lastID, anchor: .bottom)
                        }
                    }
                }
                .onAppear {
                    if let lastID = viewModel.messages.last?.id {
                        proxy.scrollTo(lastID, anchor: .bottom)
                    }
                }
            }

            if let err = viewModel.errorMessage {
                Text(err)
                    .font(.sarabun(13))
                    .foregroundStyle(Color.red.opacity(0.9))
                    .padding(.horizontal, 16)
                    .padding(.vertical, 4)
            }

            // Image attachment preview
            if let imageData = viewModel.selectedImageData, let uiImage = UIImage(data: imageData) {
                HStack(spacing: 12) {
                    Image(uiImage: uiImage)
                        .resizable()
                        .scaledToFill()
                        .frame(width: 54, height: 54)
                        .clipShape(RoundedRectangle(cornerRadius: 8))
                        .overlay(RoundedRectangle(cornerRadius: 8).stroke(Color.brandGold, lineWidth: 1))

                    Text("แนบรูปภาพแล้ว")
                        .font(.sarabun(14))
                        .foregroundStyle(Color.brandCream)

                    Spacer()

                    Button(action: handleClearAttachedImage) {
                        Image(systemName: "xmark.circle.fill")
                            .font(.system(size: 20))
                            .foregroundStyle(Color.brandCream.opacity(0.7))
                    }
                }
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
                .background(Color.brandGreen.opacity(0.4))
            }

            let hasAttachedImage = viewModel.selectedImageData != nil
            let isSending = viewModel.isSending

            // Bottom Composer bar
            VStack(spacing: 8) {
                HStack(alignment: .bottom, spacing: 10) {
                    PhotosPicker(
                        selection: $selectedPhotoItem,
                        matching: .images,
                        photoLibrary: .shared()
                    ) {
                        Image(systemName: hasAttachedImage ? "photo.fill" : "paperclip")
                            .font(.system(size: 20))
                            .foregroundStyle(hasAttachedImage ? Color.brandGold : Color.brandCream.opacity(0.8))
                            .frame(width: 40, height: 40)
                            .background(Color.brandGreen.opacity(0.4))
                            .clipShape(Circle())
                            .overlay(Circle().stroke(Color.brandGold.opacity(0.4), lineWidth: 1))
                    }
                    .task(id: selectedPhotoItem) { [viewModel] in
                        guard let selectedPhotoItem else { return }
                        if let data = try? await selectedPhotoItem.loadTransferable(type: Data.self) {
                            viewModel.attachImage(data: data)
                        }
                    }

                    TextField("พิมพ์คำถามที่อยากปรึกษา...", text: $messageText, axis: .vertical)
                        .font(.sarabun(15))
                        .lineLimit(1...5)
                        .padding(.horizontal, 12)
                        .padding(.vertical, 10)
                        .background(Color(red: 0.03, green: 0.12, blue: 0.08))
                        .clipShape(RoundedRectangle(cornerRadius: 14))
                        .overlay(RoundedRectangle(cornerRadius: 14).stroke(Color.brandGold.opacity(0.4), lineWidth: 1))
                        .foregroundStyle(Color.white)

                    Button(action: handleSendMessage) {
                        if isSending {
                            ProgressView()
                                .tint(Color.black)
                                .frame(width: 40, height: 40)
                        } else {
                            Image(systemName: "arrow.up.circle.fill")
                                .font(.system(size: 34))
                                .foregroundStyle(
                                    canSend ? Color.brandGold : Color.gray.opacity(0.5)
                                )
                        }
                    }
                    .disabled(!canSend || isSending)
                }
                .padding(.horizontal, 12)
                .padding(.vertical, 10)
                .background(Color.brandGreen.opacity(0.9))
            }
        }
    }

    private var canSend: Bool {
        !messageText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || viewModel.selectedImageData != nil
    }
}

// MARK: - Chat Bubble Row
@MainActor
private struct ChatBubbleRow: View {
    let message: SupportChatMessage
    let customerAvatarPath: String
    let onImageTap: (URL) -> Void

    var body: some View {
        let isMine = message.isFromCustomer
        let avatarPath = isMine ? customerAvatarPath : "/images/taya_avatar.jpg?v=child_hd"
        let avatarURL = APIClient.shared.avatarURL(for: avatarPath)

        HStack(alignment: .top, spacing: 8) {
            if !isMine {
                RemoteAvatarImage(url: avatarURL) {
                    Circle().fill(Color.brandGold.opacity(0.4))
                }
                .frame(width: 38, height: 38)
                .clipShape(Circle())
                .overlay(Circle().stroke(Color.brandGold.opacity(0.7), lineWidth: 1.2))
                .shadow(color: Color.black.opacity(0.3), radius: 3)
            } else {
                Spacer(minLength: 40)
            }

            VStack(alignment: isMine ? .trailing : .leading, spacing: 4) {
                HStack(spacing: 6) {
                    if isMine && (message.isRead == true) {
                        Text("อ่านแล้ว")
                            .font(.sarabun(10, weight: .regular))
                            .foregroundStyle(Color.brandCream.opacity(0.55))
                    }
                    Text(isMine ? "คุณ" : "อ.ทญา ว่องไว")
                        .font(.sarabun(12, weight: .semibold))
                        .foregroundStyle(Color.brandGold.opacity(0.9))
                }

                VStack(alignment: .leading, spacing: 8) {
                    if !message.body.isEmpty {
                        Text(message.body)
                            .font(.sarabun(15))
                            .foregroundStyle(Color.white)
                            .fixedSize(horizontal: false, vertical: true)
                    }

                    if !message.imageURL.isEmpty, let imgURL = APIClient.shared.avatarURL(for: message.imageURL) {
                        Button(action: { onImageTap(imgURL) }) {
                            AsyncImage(url: imgURL) { phase in
                                switch phase {
                                case .empty:
                                    ProgressView()
                                        .frame(height: 140)
                                        .frame(maxWidth: .infinity)
                                case .success(let image):
                                    image
                                        .resizable()
                                        .scaledToFit()
                                        .frame(maxHeight: 240)
                                        .clipShape(RoundedRectangle(cornerRadius: 10))
                                case .failure:
                                    Image(systemName: "photo")
                                        .font(.system(size: 32))
                                        .foregroundStyle(Color.gray)
                                        .frame(height: 100)
                                        .frame(maxWidth: .infinity)
                                @unknown default:
                                    EmptyView()
                                }
                            }
                        }
                        .buttonStyle(.plain)
                    }
                }
                .padding(.horizontal, 14)
                .padding(.vertical, 10)
                .background(
                    isMine ? Color(red: 0.24, green: 0.43, blue: 0.30) : Color(red: 0.34, green: 0.27, blue: 0.09)
                )
                .clipShape(
                    RoundedRectangle(cornerRadius: 16)
                )
                .shadow(color: Color.black.opacity(0.25), radius: 4, x: 0, y: 2)
            }

            if isMine {
                RemoteAvatarImage(url: avatarURL) {
                    Circle().fill(Color.brandGreen.opacity(0.6))
                }
                .frame(width: 38, height: 38)
                .clipShape(Circle())
                .overlay(Circle().stroke(Color.brandGold.opacity(0.7), lineWidth: 1.2))
                .shadow(color: Color.black.opacity(0.3), radius: 3)
            } else {
                Spacer(minLength: 40)
            }
        }
    }
}

// MARK: - Full Screen Image View
@MainActor
private struct FullScreenImageView: View {
    let url: URL
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            ZStack {
                Color.black.ignoresSafeArea()

                AsyncImage(url: url) { phase in
                    switch phase {
                    case .empty:
                        ProgressView().tint(Color.white)
                    case .success(let image):
                        image
                            .resizable()
                            .scaledToFit()
                    case .failure:
                        Text("ไม่สามารถโหลดรูปภาพได้")
                            .foregroundStyle(Color.white)
                    @unknown default:
                        EmptyView()
                    }
                }
            }
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("ปิด") { dismiss() }
                        .foregroundStyle(Color.brandGold)
                }
            }
        }
    }
}
