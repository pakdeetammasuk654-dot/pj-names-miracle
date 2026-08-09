package com.miracle.names.ui.components

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.miracle.names.data.model.PairDetailResponse

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PairDetailSheet(
    pairDetail: PairDetailResponse?,
    isLoading: Boolean,
    onDismiss: () -> Unit
) {
    if (pairDetail == null && !isLoading) return

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true),
        containerColor = Color(0xFFF0F7F4),
        dragHandle = {
            BottomSheetDefaults.DragHandle(
                color = Color(0xFF74C69D)
            )
        }
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .verticalScroll(rememberScrollState())
                .padding(horizontal = 24.dp, vertical = 12.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            if (isLoading && pairDetail == null) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(180.dp),
                    contentAlignment = Alignment.Center
                ) {
                    CircularProgressIndicator(color = Color(0xFF2D6A4F))
                }
            } else if (pairDetail != null) {
                // Header with Badge Circle
                val isGood = pairDetail.isGood
                val badgeBgColor = when {
                    !isGood -> Color(0xFFC62828)
                    pairDetail.pairType.contains("D10") -> Color(0xFF1B5E20)
                    pairDetail.pairType.contains("D8") || pairDetail.pairType == "D" -> Color(0xFF2E7D32)
                    pairDetail.pairType.contains("D5") -> Color(0xFF4CAF50)
                    else -> Color(0xFFE57373)
                }

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(10.dp)
                    ) {
                        Text(
                            text = "คู่เลข",
                            color = Color(0xFF1B4332),
                            fontWeight = FontWeight.Bold,
                            fontSize = 18.sp
                        )

                        Box(
                            modifier = Modifier
                                .size(40.dp)
                                .clip(CircleShape)
                                .background(badgeBgColor),
                            contentAlignment = Alignment.Center
                        ) {
                            Text(
                                text = pairDetail.pairNumber,
                                color = Color.White,
                                fontSize = 18.sp,
                                fontWeight = FontWeight.Bold
                            )
                        }
                    }

                    IconButton(onClick = onDismiss) {
                        Icon(
                            imageVector = Icons.Default.Close,
                            contentDescription = "ปิด",
                            tint = Color(0xFF2D6A4F)
                        )
                    }
                }

                HorizontalDivider(color = Color(0xFFB7E4C7))

                // Short Summary Card (MiracleDesc)
                if (pairDetail.miracleDesc.isNotBlank()) {
                    Surface(
                        modifier = Modifier.fillMaxWidth(),
                        shape = RoundedCornerShape(14.dp),
                        color = Color(0xFFD8F3DC),
                        border = BorderStroke(1.dp, Color(0xFFB7E4C7))
                    ) {
                        Row(
                            modifier = Modifier.padding(14.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Icon(
                                imageVector = Icons.Default.AutoAwesome,
                                contentDescription = null,
                                tint = Color(0xFF2D6A4F),
                                modifier = Modifier.size(20.dp)
                            )
                            Spacer(modifier = Modifier.width(10.dp))
                            Text(
                                text = pairDetail.miracleDesc,
                                color = Color(0xFF1B4332),
                                fontWeight = FontWeight.Bold,
                                fontSize = 14.sp
                            )
                        }
                    }
                }

                // Detailed Meaning (MiracleDetail - ความหมายโดยละเอียดมาก)
                val detailText = pairDetail.miracleDetail.ifBlank {
                    pairDetail.miracleDesc.ifBlank { "คู่เลข ${pairDetail.pairNumber} เป็นคู่เลขทรงพลังตามตำราเลขศาสตร์ไทย" }
                }

                Surface(
                    color = Color(0xFFE8F5E9),
                    shape = RoundedCornerShape(14.dp),
                    border = BorderStroke(1.dp, Color(0xFFC8E6C9)),
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Text(
                        text = detailText,
                        color = Color(0xFF1B4332),
                        fontSize = 13.sp,
                        lineHeight = 22.sp,
                        modifier = Modifier.padding(16.dp)
                    )
                }

                Spacer(modifier = Modifier.height(16.dp))
            }
        }
    }
}
