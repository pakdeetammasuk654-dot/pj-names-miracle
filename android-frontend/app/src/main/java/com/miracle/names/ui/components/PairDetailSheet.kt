package com.miracle.names.ui.components

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.miracle.names.data.model.PairDetailResponse
import com.miracle.names.ui.results.PairCircles

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
        containerColor = Color(0xFF071C14),
        dragHandle = {
            BottomSheetDefaults.DragHandle(
                color = Color(0xFFD0A12B).copy(alpha = 0.78f)
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
                    CircularProgressIndicator(color = Color(0xFFD0A12B))
                }
            } else if (pairDetail != null) {
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
                            color = Color(0xFFF7E7B0),
                            fontWeight = FontWeight.Bold,
                            fontSize = 18.sp
                        )

                        PairCircles(
                            pairsStr = pairDetail.pairNumber,
                            typesStr = pairDetail.pairType,
                            goodFlags = listOf(pairDetail.isGood),
                            circleSize = 34.dp,
                            fontSize = 14.sp,
                            strokeWidth = 2.dp
                        )
                    }

                    IconButton(onClick = onDismiss) {
                        Icon(
                            imageVector = Icons.Default.Close,
                            contentDescription = "ปิด",
                            tint = Color(0xFFF7E7B0).copy(alpha = 0.82f)
                        )
                    }
                }

                HorizontalDivider(color = Color(0xFFD0A12B).copy(alpha = 0.42f))

                // Short Summary Card (MiracleDesc)
                if (pairDetail.miracleDesc.isNotBlank()) {
                    Surface(
                        modifier = Modifier.fillMaxWidth(),
                        shape = RoundedCornerShape(14.dp),
                        color = Color(0xFF0E3425),
                        border = BorderStroke(1.dp, Color(0xFFD0A12B).copy(alpha = 0.34f))
                    ) {
                        Row(
                            modifier = Modifier.padding(14.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Icon(
                                imageVector = Icons.Default.AutoAwesome,
                                contentDescription = null,
                                tint = Color(0xFFD0A12B),
                                modifier = Modifier.size(20.dp)
                            )
                            Spacer(modifier = Modifier.width(10.dp))
                            Text(
                                text = pairDetail.miracleDesc,
                                color = Color(0xFFF7E7B0),
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
                    color = Color(0xFF10251C),
                    shape = RoundedCornerShape(14.dp),
                    border = BorderStroke(1.dp, Color(0xFFD0A12B).copy(alpha = 0.24f)),
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Text(
                        text = detailText,
                        color = Color(0xFFEDE5D1),
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
