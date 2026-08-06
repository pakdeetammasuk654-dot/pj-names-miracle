package com.miracle.names.ui.blessing

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.filled.Share
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.layout.ContentScale
import coil.compose.AsyncImage
import coil.ImageLoader
import coil.decode.SvgDecoder
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.miracle.names.data.model.NameResultItem
import com.miracle.names.ui.theme.*
import com.miracle.names.viewmodel.NameSearchViewModel
import android.widget.Toast
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.graphics.asAndroidBitmap
import androidx.compose.ui.graphics.Brush
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.ExperimentalComposeApi
import kotlinx.coroutines.launch
import dev.shreyaspatil.capturable.capturable
import dev.shreyaspatil.capturable.controller.rememberCaptureController
import com.miracle.names.utils.ImageUtils
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue

@OptIn(ExperimentalMaterial3Api::class, ExperimentalComposeUiApi::class, ExperimentalComposeApi::class)
@Composable
fun BlessingCardScreen(
    item: NameResultItem? = null,
    viewModel: NameSearchViewModel,
    onBackClick: () -> Unit = { viewModel.selectNameForBlessingCard(null) }
) {
    val selectedItem = item ?: viewModel.selectedItem.collectAsState().value
    val dayOfBirth by viewModel.dayOfBirth.collectAsState()
    val surname by viewModel.surname.collectAsState()
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    val captureController = rememberCaptureController()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("การ์ดสิริมงคล", color = GoldLight) },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(
                            imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                            contentDescription = "กลับ",
                            tint = GoldLight
                        )
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = DarkBackground)
            )
        },
        containerColor = DarkBackground
    ) { paddingValues ->
        if (selectedItem == null) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(paddingValues),
                contentAlignment = Alignment.Center
            ) {
                Text("ไม่พบข้อมูลการ์ดมงคล", color = Color.White)
            }
        } else {
            val selectedName = selectedItem
            val fullName = if (surname.isNotBlank()) "${selectedName.thName} $surname" else selectedName.thName

            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(paddingValues)
                    .verticalScroll(rememberScrollState())
                    .padding(16.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Top
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .aspectRatio(9f/16f)
                        .capturable(captureController)
                            .clip(RoundedCornerShape(24.dp))
                            .background(
                                brush = Brush.verticalGradient(
                                    colors = listOf(Color(0xFF1E0A3C), Color(0xFF0F0524), Color(0xFF1E0A3C))
                                ),
                                shape = RoundedCornerShape(24.dp)
                            )
                            .border(width = 3.dp, color = GoldPrimary, shape = RoundedCornerShape(24.dp))
                            .padding(2.dp)
                            .border(width = 1.dp, color = GoldLight, shape = RoundedCornerShape(22.dp))
                            .padding(16.dp),
                        contentAlignment = Alignment.Center
                    ) {
                        val svgImageLoader = ImageLoader.Builder(context)
                            .components {
                                add(SvgDecoder.Factory())
                            }
                            .build()

                        AsyncImage(
                            model = "http://10.0.2.2:8080/buddhas/${dayOfBirth.lowercase()}_buddha.svg",
                            imageLoader = svgImageLoader,
                            contentDescription = "พระประจำวัน${formatDayThai(dayOfBirth)}",
                            modifier = Modifier.fillMaxSize(),
                            contentScale = ContentScale.Fit,
                            alpha = 0.15f
                        )

                        Column(
                            horizontalAlignment = Alignment.CenterHorizontally,
                            verticalArrangement = Arrangement.Center
                        ) {
                            // Person Name Display
                            Text(
                                text = fullName,
                                color = GoldPrimary,
                                fontWeight = FontWeight.Bold,
                                fontSize = 24.sp,
                                textAlign = TextAlign.Center
                            )

                            Spacer(modifier = Modifier.height(12.dp))

                            // Meaning Display
                            Text(
                                text = selectedName.meaning.ifEmpty { "ผู้มีความสุขความเจริญ สติปัญญา และมั่งคั่งสมบูรณ์พร้อม" },
                                color = Color.White,
                                fontSize = 14.sp,
                                textAlign = TextAlign.Center
                            )

                            HorizontalDivider(
                                modifier = Modifier.padding(vertical = 12.dp),
                                color = Color.Gray.copy(alpha = 0.3f)
                            )

                            // Day-of-birth Kalakini safety information
                            Column(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalAlignment = Alignment.CenterHorizontally,
                                verticalArrangement = Arrangement.spacedBy(6.dp)
                            ) {
                                Text(
                                    text = "เกิด${formatDayThai(dayOfBirth)} ปลอดกาลกิณี",
                                    color = GreenBadgeText,
                                    fontSize = 13.sp,
                                    textAlign = TextAlign.Center
                                )
                            }

                            HorizontalDivider(
                                modifier = Modifier.padding(vertical = 12.dp),
                                color = Color.Gray.copy(alpha = 0.3f)
                            )

                            val satDescDisplay = if (selectedName.satMiracleDesc.isNotEmpty()) selectedName.satMiracleDesc else "ถอดคู่เลขมงคลสมบูรณ์"
                            val satPairsText = if (selectedName.satPairsDisplay.isNotEmpty()) selectedName.satPairsDisplay else "${selectedName.satSum}"

                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(vertical = 3.dp),
                                horizontalArrangement = Arrangement.Center,
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Text(
                                    text = "• เลขศาสตร์: ",
                                    color = Color.White,
                                    fontSize = 12.sp
                                )
                                com.miracle.names.ui.results.PairCircles(pairsStr = satPairsText, typesStr = selectedName.satPairType)
                                Spacer(modifier = Modifier.width(6.dp))
                                Text(
                                    text = "➔ ${satDescDisplay}",
                                    color = Color.White,
                                    fontSize = 12.sp,
                                    modifier = Modifier.weight(1f, fill = false),
                                    textAlign = TextAlign.Center
                                )
                            }

                            val shaDescDisplay = if (selectedName.shaMiracleDesc.isNotEmpty()) selectedName.shaMiracleDesc else "พลังเลขเงาสมบูรณ์"
                            val shaPairsText = if (selectedName.shaPairsDisplay.isNotEmpty()) selectedName.shaPairsDisplay else "${selectedName.shaSum}"

                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(vertical = 3.dp),
                                horizontalArrangement = Arrangement.Center,
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Text(
                                    text = "• พลังเงา: ",
                                    color = GoldLight,
                                    fontSize = 12.sp
                                )
                                com.miracle.names.ui.results.PairCircles(pairsStr = shaPairsText, typesStr = selectedName.shaPairType)
                                Spacer(modifier = Modifier.width(6.dp))
                                Text(
                                    text = "➔ ${shaDescDisplay}",
                                    color = GoldLight,
                                    fontSize = 12.sp,
                                    modifier = Modifier.weight(1f, fill = false),
                                    textAlign = TextAlign.Center
                                )
                            }

                            if (surname.isNotEmpty() && selectedName.totalShaSum > 0) {
                                val totalShaDescDisplay = if (selectedName.shaMiracleDesc.isNotEmpty()) selectedName.shaMiracleDesc else "ถอดคู่เลขมงคลรวมนามสกุลสมบูรณ์"
                                val totalShaPairsText = if (selectedName.shaPairsDisplay.isNotEmpty()) selectedName.shaPairsDisplay else "${selectedName.totalShaSum}"

                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(vertical = 3.dp),
                                    horizontalArrangement = Arrangement.Center,
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    Text(
                                        text = "• รวมนามสกุล: ",
                                        color = Color.White,
                                        fontSize = 12.sp
                                    )
                                    com.miracle.names.ui.results.PairCircles(pairsStr = totalShaPairsText, typesStr = selectedName.shaPairType)
                                    Spacer(modifier = Modifier.width(6.dp))
                                    Text(
                                        text = "➔ ${totalShaDescDisplay}",
                                        color = Color.White,
                                        fontSize = 12.sp,
                                        modifier = Modifier.weight(1f, fill = false),
                                        textAlign = TextAlign.Center
                                    )
                                }
                            }
                        }
                    }
                Spacer(modifier = Modifier.height(16.dp))

                // Action Buttons
                Button(
                    onClick = {
                        scope.launch {
                            try {
                                val bitmap = captureController.captureAsync().await()
                                val success = ImageUtils.saveBitmapToGallery(context, bitmap.asAndroidBitmap())
                                if (success) {
                                    Toast.makeText(context, "บันทึกรูปลงเครื่องสำเร็จ!", Toast.LENGTH_SHORT).show()
                                } else {
                                    Toast.makeText(context, "เกิดข้อผิดพลาดในการบันทึกรูปภาพ", Toast.LENGTH_SHORT).show()
                                }
                            } catch (_: Throwable) {
                                Toast.makeText(context, "ไม่สามารถจับภาพหน้าจอได้", Toast.LENGTH_SHORT).show()
                            }
                        }
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(50.dp),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = GoldPrimary)
                ) {
                    Icon(
                        imageVector = Icons.Default.Share,
                        contentDescription = null,
                        tint = Color.Black
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "แชร์ / บันทึกการ์ดมงคลวิเคราะห์นาม",
                        color = Color.Black,
                        fontWeight = FontWeight.Bold
                    )
                }
            }
        }
    }
}

private fun formatDayThai(day: String): String {
    return when (day.lowercase()) {
        "sunday" -> "วันอาทิตย์"
        "monday" -> "วันจันทร์"
        "tuesday" -> "วันอังคาร"
        "wednesday" -> "วันพุธ (กลางวัน)"
        "wednesday_night" -> "วันพุธ (กลางคืน)"
        "thursday" -> "วันพฤหัสบดี"
        "friday" -> "วันศุกร์"
        "saturday" -> "วันเสาร์"
        else -> day
    }
}
