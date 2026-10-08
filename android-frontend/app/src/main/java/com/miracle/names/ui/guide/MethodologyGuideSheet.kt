package com.miracle.names.ui.guide

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Language
import androidx.compose.foundation.clickable
import androidx.compose.material3.*
import androidx.compose.material3.TabRowDefaults.tabIndicatorOffset
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil.compose.AsyncImage
import com.miracle.names.data.model.ArticleItem
import com.miracle.names.ui.theme.GoldLight
import com.miracle.names.ui.theme.GoldPrimary
import com.miracle.names.ui.theme.GreenHeader

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MethodologyGuideSheet(
    articles: List<ArticleItem> = emptyList(),
    onDismiss: () -> Unit
) {
    val uriHandler = LocalUriHandler.current
    val scrollState = rememberScrollState()

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = false),
        containerColor = GreenHeader
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .verticalScroll(scrollState)
                .padding(horizontal = 20.dp, vertical = 12.dp)
                .padding(bottom = 32.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            // Header Row
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "✨ แนะนำ อ.ทญา ว่องไว",
                    color = GoldPrimary,
                    fontWeight = FontWeight.Bold,
                    fontSize = 20.sp,
                    modifier = Modifier.weight(1f)
                )

                IconButton(
                    onClick = onDismiss,
                    modifier = Modifier.size(32.dp)
                ) {
                    Icon(
                        imageVector = Icons.Filled.Close,
                        contentDescription = "ปิด",
                        tint = GoldPrimary,
                        modifier = Modifier.size(20.dp)
                    )
                }
            }

            // Photo of Ajarn Taya
            Column(
                modifier = Modifier.fillMaxWidth(),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(6.dp)
            ) {
                AsyncImage(
                    model = "https://xn--b3cu8e7ah6h.com/images/taya_photo.jpg?v=5",
                    contentDescription = "อ.ทญา ว่องไว ในวัยประถมศึกษา",
                    contentScale = ContentScale.Crop,
                    modifier = Modifier
                        .size(width = 150.dp, height = 175.dp)
                        .clip(RoundedCornerShape(16.dp))
                        .border(1.5.dp, GoldPrimary.copy(alpha = 0.6f), RoundedCornerShape(16.dp))
                )
                Text(
                    text = "อ.ทญา ว่องไว ในวัยประถมศึกษา",
                    color = GoldLight,
                    fontWeight = FontWeight.SemiBold,
                    fontSize = 12.sp
                )
                Text(
                    text = "จุดเริ่มต้นและแรงบันดาลใจในการศึกษาค้นคว้าศาสตร์แห่งชื่อมงคล",
                    color = Color.White.copy(alpha = 0.68f),
                    fontSize = 11.sp,
                    textAlign = TextAlign.Center
                )
            }

            HorizontalDivider(color = GoldPrimary.copy(alpha = 0.35f))

            // Section 1: Background & Inspiration
            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Text(
                    text = "ความเป็นมาและความสำคัญของปัญหา",
                    color = GoldPrimary,
                    fontWeight = FontWeight.Bold,
                    fontSize = 15.sp
                )

                Text(
                    text = "สมัยเรียนชั้นประถมศึกษา ครูประจำชั้นจะขานชื่อนักเรียนทุกเช้า เพื่อนบางคนมีชื่อและบุคลิกแปลกชวนขบขัน เช่น ครูเรียกชื่อ “เด็กชายมังกือ” ก็จะมีเสียงหัวเราะครึกครื้นในห้อง",
                    color = Color.White.copy(alpha = 0.9f),
                    fontSize = 13.sp,
                    lineHeight = 19.sp
                )

                Text(
                    text = "มีบางครั้งที่ชื่อของเพื่อนเลขที่ติดกัน เมื่อออกเสียงต่อกันแล้วฟังคล้ายบทสวดมนต์ จนคุณครูแซวว่า:",
                    color = Color.White.copy(alpha = 0.9f),
                    fontSize = 13.sp,
                    lineHeight = 19.sp
                )

                // Quote Box
                Surface(
                    shape = RoundedCornerShape(8.dp),
                    color = GoldPrimary.copy(alpha = 0.10f),
                    border = BorderStroke(1.dp, GoldPrimary.copy(alpha = 0.35f)),
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Row(
                        modifier = Modifier.padding(12.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Box(
                            modifier = Modifier
                                .width(4.dp)
                                .height(32.dp)
                                .background(GoldPrimary, RoundedCornerShape(2.dp))
                        )
                        Spacer(modifier = Modifier.width(10.dp))
                        Text(
                            text = "“ชื่อแบบนี้เอามาเขียนต่อกันแล้วสวดเป่าแก้ตาแดงได้”",
                            color = GoldLight,
                            fontWeight = FontWeight.SemiBold,
                            fontSize = 13.sp,
                            lineHeight = 18.sp
                        )
                    }
                }

                Text(
                    text = "คำพูดเปรียบเปรยของคุณครูในวันนั้น ที่เชื่อมโยงบทสวดมนต์ ภูมิปัญญาโบราณ และการรักษาโรค เข้ากับ 'ชื่อ' ได้กลายเป็นจุดประกายและแรงบันดาลใจให้ค้นคว้าอย่างลึกซึ้ง: เราจะตั้งชื่ออย่างไรให้ถูกต้องตามตำราและส่งเสริมชีวิตได้จริง?",
                    color = Color.White.copy(alpha = 0.9f),
                    fontSize = 13.sp,
                    lineHeight = 19.sp
                )
            }

            // Section 2: Author & Data Science
            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Text(
                    text = "เกี่ยวกับผู้เขียนและผู้พัฒนาระบบ",
                    color = GoldPrimary,
                    fontWeight = FontWeight.Bold,
                    fontSize = 15.sp
                )

                Text(
                    text = "แอป “ชื่อดี” พัฒนาขึ้นเมื่อองค์ความรู้ด้านการออกแบบชื่อตกผลึก และความสามารถของเทคโนโลยี AI มีความพร้อม โดย ทญา ว่องไว นำหลัก Data Science มาจัดโครงสร้างการสืบค้นและเปรียบเทียบชื่อ ผสานกับ 4 มิติสำคัญ:",
                    color = Color.White.copy(alpha = 0.9f),
                    fontSize = 13.sp,
                    lineHeight = 19.sp
                )

                Surface(
                    shape = RoundedCornerShape(12.dp),
                    color = Color(0xFF1F120B).copy(alpha = 0.90f),
                    border = BorderStroke(1.dp, GoldPrimary.copy(alpha = 0.30f)),
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Column(
                        modifier = Modifier.padding(12.dp),
                        verticalArrangement = Arrangement.spacedBy(6.dp)
                    ) {
                        DimensionRowItem("⭐ เลขศาสตร์ & พลังเงา", "คำนวณคู่เลขมงคลและผลรวมตามตำราชั้นสูงอย่างแม่นยำ")
                        DimensionRowItem("🎯 ความหมายมงคล", "วิเคราะห์ความหมายลึกซึ้ง ตรงจุดประสงค์ เสริมมงคลชีวิต")
                        DimensionRowItem("🎶 ความไพเราะของเสียง", "ลื่นไหล จังหวะสัมผัส และพลังเสียงที่เป็นสิริมงคล")
                        DimensionRowItem("🔍 เอกลักษณ์เฉพาะตัว", "คงรากศัพท์และกลิ่นอายเสียงเดิมของชื่อต้นแบบที่คุณชื่นชอบ")
                    }
                }

                Text(
                    text = "“การเลือกใช้ชื่อที่ดี จะแสดงรสนิยม ส่งเสริมภาพลักษณ์ และมอบประสบการณ์ที่ดีแก่ชีวิต”",
                    color = Color.White.copy(alpha = 0.85f),
                    fontSize = 12.sp,
                    fontStyle = FontStyle.Italic,
                    lineHeight = 18.sp
                )

                Text(
                    text = "— ทญา ว่องไว",
                    color = GoldLight,
                    fontWeight = FontWeight.Bold,
                    fontSize = 13.sp,
                    textAlign = TextAlign.End,
                    modifier = Modifier.fillMaxWidth()
                )
            }

            HorizontalDivider(color = GoldPrimary.copy(alpha = 0.35f))

            // CTA Button to website landing page
            Button(
                onClick = { uriHandler.openUri("https://xn--b3cu8e7ah6h.com/") },
                colors = ButtonDefaults.buttonColors(containerColor = GoldPrimary),
                shape = RoundedCornerShape(12.dp),
                modifier = Modifier
                    .fillMaxWidth()
                    .height(48.dp)
            ) {
                Icon(
                    imageVector = Icons.Default.Language,
                    contentDescription = null,
                    tint = Color(0xFF0F3524),
                    modifier = Modifier.size(18.dp)
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = "เข้าสู่หน้าเว็บไซต์หลัก (ชื่อดี.com)",
                    color = Color(0xFF0F3524),
                    fontWeight = FontWeight.Bold,
                    fontSize = 14.sp
                )
            }
        }
    }
}

@Composable
private fun DimensionRowItem(title: String, desc: String) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.Top
    ) {
        Text(
            text = "•",
            color = GoldPrimary,
            fontSize = 12.sp,
            modifier = Modifier.padding(end = 6.dp)
        )
        Column(verticalArrangement = Arrangement.spacedBy(1.dp)) {
            Text(
                text = title,
                color = GoldLight,
                fontWeight = FontWeight.Bold,
                fontSize = 12.sp
            )
            Text(
                text = desc,
                color = Color.White.copy(alpha = 0.80f),
                fontSize = 11.sp,
                lineHeight = 15.sp
            )
        }
    }
}
