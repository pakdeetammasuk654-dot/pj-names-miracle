import SwiftUI
import PhotosUI

struct AdminChatView: View {
    let conversation: AdminConversation
    @StateObject private var viewModel: AdminChatViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var messageText = ""
    @State private var selectedPhotoItem: PhotosPickerItem?
    @State private var selectedImageData: Data?
    @State private var selectedImageMimeType: String?
    @State private var fullScreenImageURL: URL?
    @State private var isFullScreenPresented = false
    @State private var messagePendingDeletion: AdminMessage?

    init(conversation: AdminConversation) {
        self.conversation = conversation
        _viewModel = StateObject(wrappedValue: AdminChatViewModel(conversation: conversation))
    }

    var body: some View {
        ZStack {
            Color.brandBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                // Messages List
                ScrollViewReader { proxy in
                    ScrollView {
                        LazyVStack(spacing: 14) {
                            ForEach(viewModel.messages) { msg in
                                AdminMessageBubble(
                                    message: msg,
                                    customerAvatarKey: viewModel.conversation.customerAvatar ?? conversation.customerAvatar,
                                    customerName: viewModel.conversation.customerName ?? conversation.customerName,
                                    onImageTap: { url in
                                        fullScreenImageURL = url
                                        isFullScreenPresented = true
                                    }
                                )
                                .contextMenu {
                                    Button(role: .destructive) {
                                        messagePendingDeletion = msg
                                    } label: {
                                        Label("ลบข้อความ", systemImage: "trash")
                                    }
                                }
                                .id(msg.id)
                            }
                        }
                        .padding(.horizontal, 16)
                        .padding(.vertical, 14)
                    }
                    .onChange(of: viewModel.messages.count) { _ in
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
                        .foregroundStyle(Color.brandRed)
                        .padding(.horizontal, 16)
                        .padding(.vertical, 4)
                }

                // Image Preview if attached
                if let data = selectedImageData, let uiImage = UIImage(data: data) {
                    HStack(spacing: 12) {
                        Image(uiImage: uiImage)
                            .resizable()
                            .scaledToFill()
                            .frame(width: 50, height: 50)
                            .clipShape(RoundedRectangle(cornerRadius: 8))
                            .overlay(RoundedRectangle(cornerRadius: 8).stroke(Color.brandGold, lineWidth: 1))

                        Text("แนบรูปภาพแล้ว")
                            .font(.sarabun(13))
                            .foregroundStyle(Color.brandCream)

                        Spacer()

                        Button(action: {
                            selectedPhotoItem = nil
                            selectedImageData = nil
                            selectedImageMimeType = nil
                        }) {
                            Image(systemName: "xmark.circle.fill")
                                .font(.system(size: 18))
                                .foregroundStyle(Color.brandCream.opacity(0.7))
                        }
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 6)
                    .background(Color.brandGreen.opacity(0.4))
                }

                // Composer
                VStack(spacing: 6) {
                    HStack(alignment: .bottom, spacing: 10) {
                        PhotosPicker(
                            selection: $selectedPhotoItem,
                            matching: .images,
                            photoLibrary: .shared()
                        ) {
                            Image(systemName: selectedImageData == nil ? "paperclip" : "photo.fill")
                                .font(.system(size: 18))
                                .foregroundStyle(selectedImageData == nil ? Color.brandCream.opacity(0.8) : Color.brandGold)
                                .frame(width: 38, height: 38)
                                .background(Color.brandGreen.opacity(0.4))
                                .clipShape(Circle())
                                .overlay(Circle().stroke(Color.brandGold.opacity(0.3), lineWidth: 1))
                        }
                        .onChange(of: selectedPhotoItem) { newItem in
                            Task {
                                if let data = try? await newItem?.loadTransferable(type: Data.self) {
                                    selectedImageData = data
                                    selectedImageMimeType = "image/jpeg"
                                }
                            }
                        }

                        TextField("พิมพ์ตอบลูกค้าในนาม อ.ทญา...", text: $messageText, axis: .vertical)
                            .font(.sarabun(15))
                            .lineLimit(1...5)
                            .padding(.horizontal, 12)
                            .padding(.vertical, 8)
                            .background(Color(red: 0.03, green: 0.12, blue: 0.08))
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                            .overlay(RoundedRectangle(cornerRadius: 12).stroke(Color.brandGold.opacity(0.4), lineWidth: 1))
                            .foregroundStyle(Color.white)

                        Button(action: {
                            let textToSend = messageText
                            let imgToSend = selectedImageData
                            let mimeToSend = selectedImageMimeType

                            messageText = ""
                            selectedPhotoItem = nil
                            selectedImageData = nil
                            selectedImageMimeType = nil

                            Task {
                                await viewModel.sendMessage(
                                    body: textToSend,
                                    imageData: imgToSend,
                                    mimeType: mimeToSend
                                )
                            }
                        }) {
                            if viewModel.isSending {
                                ProgressView().tint(.black).frame(width: 36, height: 36)
                            } else {
                                Image(systemName: "arrow.up.circle.fill")
                                    .font(.system(size: 32))
                                    .foregroundStyle(canSend ? Color.brandGold : Color.gray.opacity(0.4))
                            }
                        }
                        .disabled(!canSend || viewModel.isSending)
                    }
                    .padding(.horizontal, 12)
                    .padding(.vertical, 8)
                    .background(Color.brandGreen.opacity(0.95))
                }
            }
        }
        .navigationTitle(viewModel.conversation.displayName)
        .navigationBarTitleDisplayMode(.inline)
        .toolbarBackground(Color.brandGreen.opacity(0.95), for: .navigationBar)
        .toolbarBackground(.visible, for: .navigationBar)
        .toolbarColorScheme(.dark, for: .navigationBar)
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                if viewModel.conversation.status == "open" {
                    Button("ปิดเคส") {
                        Task {
                            await viewModel.toggleCaseStatus()
                            dismiss()
                        }
                    }
                    .font(.sarabun(13, weight: .bold))
                    .foregroundStyle(Color.brandGold)
                }
            }
        }
        .sheet(isPresented: $isFullScreenPresented) {
            if let url = fullScreenImageURL {
                FullScreenImageModal(url: url)
            }
        }
        .alert("ลบข้อความนี้?", isPresented: Binding(
            get: { messagePendingDeletion != nil },
            set: { if !$0 { messagePendingDeletion = nil } }
        ), presenting: messagePendingDeletion) { message in
            Button("ลบ", role: .destructive) {
                Task { await viewModel.deleteMessage(message) }
            }
            Button("ยกเลิก", role: .cancel) {}
        } message: { _ in
            Text("ข้อความจะหายจากบทสนทนาของลูกค้าและอ.ทญา และไม่สามารถกู้คืนได้")
        }
        .task {
            await viewModel.loadMessages()
            viewModel.startPolling()
        }
        .onDisappear {
            viewModel.stopPolling()
        }
    }

    private var canSend: Bool {
        !messageText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || selectedImageData != nil
    }
}
