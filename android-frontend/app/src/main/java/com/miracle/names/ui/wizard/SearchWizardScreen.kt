package com.miracle.names.ui.wizard

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.animateScrollBy
import androidx.compose.foundation.gestures.awaitEachGesture
import androidx.compose.foundation.gestures.awaitFirstDown
import androidx.compose.foundation.gestures.detectDragGestures
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.animation.togetherWith
import coil.compose.AsyncImage
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.layout.onGloballyPositioned
import androidx.compose.ui.layout.positionInParent
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.drawWithContent
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.res.painterResource
import com.miracle.names.R
import com.miracle.names.ui.theme.GoldPrimary
import com.miracle.names.ui.theme.PromptFontFamily

import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.scaleIn
import androidx.compose.animation.scaleOut
import androidx.compose.animation.expandVertically
import androidx.compose.animation.shrinkVertically
import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.tween
import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.spring
import androidx.compose.animation.core.LinearOutSlowInEasing
import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.RepeatMode
import androidx.compose.ui.graphics.graphicsLayer
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlin.math.roundToInt
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.automirrored.filled.LibraryBooks
import androidx.compose.material.icons.automirrored.filled.MenuBook
import androidx.compose.material.icons.filled.Clear
import androidx.compose.material.icons.filled.CalendarMonth
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.EditNote
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.automirrored.filled.Chat
import androidx.compose.material.icons.filled.BugReport
import androidx.compose.material.icons.filled.Lock
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Shield
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material.icons.filled.Info
import androidx.compose.material3.*
import androidx.compose.material3.Slider
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.draw.clipToBounds
import androidx.compose.ui.input.nestedscroll.NestedScrollConnection
import androidx.compose.ui.input.nestedscroll.NestedScrollSource
import androidx.compose.ui.input.nestedscroll.nestedScroll
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.layout
import androidx.compose.ui.platform.LocalUriHandler
import com.miracle.names.data.model.ArticleItem
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.text.SpanStyle
import androidx.compose.foundation.BorderStroke
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.miracle.names.ui.theme.DarkColorScheme
import com.miracle.names.ui.theme.GoldPrimary
import com.miracle.names.ui.theme.GoldLight
import com.miracle.names.ui.theme.GreenHeader
import com.miracle.names.ui.theme.PurpleHeader
import com.miracle.names.ui.theme.PurpleDark
import com.miracle.names.ui.theme.TextSecondary
import com.miracle.names.ui.theme.CardBgDark
import com.miracle.names.ui.results.PairCircles
import com.miracle.names.ui.components.RadarWeightSelector
import androidx.compose.foundation.shape.CircleShape
import com.miracle.names.data.remote.MiracleApiConfig
import com.miracle.names.viewmodel.NameSearchViewModel

@Composable
fun ModeSegmentedControl(
    selectedMode: String,
    onModeSelected: (String) -> Unit,
    modifier: Modifier = Modifier
) {
    Surface(
        shape = RoundedCornerShape(12.dp),
        color = Color(0xFF160C08).copy(alpha = 0.85f),
        border = BorderStroke(1.dp, Color.White.copy(alpha = 0.12f)),
        modifier = modifier.fillMaxWidth()
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(5.dp),
            horizontalArrangement = Arrangement.spacedBy(6.dp)
        ) {
            val options = listOf(
                "standalone" to "เฉพาะชื่อตัว",
                "combined" to "ชื่อ + นามสกุล"
            )

            options.forEach { (modeKey, label) ->
                val isSelected = selectedMode == modeKey
                val itemShape = RoundedCornerShape(10.dp)
                val activeTextColor = Color(0xFF1F120B)
                val inactiveTextColor = GoldLight.copy(alpha = 0.82f)

                Box(
                    modifier = Modifier
                        .weight(1f)
                        .clip(itemShape)
                        .then(
                            if (isSelected) {
                                Modifier.background(
                                    brush = Brush.horizontalGradient(
                                        colors = listOf(
                                            Color(0xFFD0A12B),
                                            Color(0xFFE6C85A),
                                            Color(0xFFD0A12B)
                                        )
                                    )
                                )
                            } else {
                                Modifier
                                    .background(Color(0xFF26160E).copy(alpha = 0.72f))
                                    .border(
                                        width = 1.dp,
                                        color = GoldPrimary.copy(alpha = 0.18f),
                                        shape = itemShape
                                    )
                            }
                        )
                        .clickable { onModeSelected(modeKey) }
                        .padding(vertical = 10.dp),
                    contentAlignment = Alignment.Center
                ) {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.Center
                    ) {
                        Icon(
                            imageVector = if (modeKey == "standalone") Icons.Default.Person else Icons.AutoMirrored.Filled.MenuBook,
                            contentDescription = null,
                            tint = if (isSelected) activeTextColor else inactiveTextColor,
                            modifier = Modifier
                                .size(18.dp)
                                .padding(end = 4.dp)
                        )
                        Text(
                            text = label,
                            fontSize = 13.sp,
                            fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Medium,
                            color = if (isSelected) activeTextColor else inactiveTextColor
                        )
                    }
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SearchWizardScreen(
    viewModel: NameSearchViewModel,
    onOpenChat: () -> Unit = {},
    onOpenAuspicious: () -> Unit = {},
    unreadCount: Int = 0,
    onOpenDebugPurchase: (() -> Unit)? = null
) {
    MaterialTheme(colorScheme = DarkColorScheme) {
        val searchMode by viewModel.searchMode.collectAsState()
        val isFindSurnameMode by viewModel.isFindSurnameMode.collectAsState()
        val surname by viewModel.surname.collectAsState()
        val dayOfBirth by viewModel.dayOfBirth.collectAsState()
        val prototypeName by viewModel.prototypeName.collectAsState()
        val nameSuggestions by viewModel.nameSuggestions.collectAsState()
        val suggestionPreviewPairs by viewModel.suggestionPreviewPairs.collectAsState()
        val suggestionCount by viewModel.suggestionCount.collectAsState()
        val suggestionsUnlocked by viewModel.suggestionsUnlocked.collectAsState()
        val isUnlocked by viewModel.isUnlocked.collectAsState()
        val isLoadingNameSuggestions by viewModel.isLoadingNameSuggestions.collectAsState()
        val selectedSuggestionName by viewModel.selectedSuggestionName.collectAsState()
        val selectedWish by viewModel.selectedWish.collectAsState()
        val pendingWeights by viewModel.pendingRankingWeights.collectAsState()
        val celebrities by viewModel.celebrities.collectAsState()
        val decodedPairs by viewModel.decodedPairs.collectAsState()
        val decodedPairTypes by viewModel.decodedPairTypes.collectAsState()
        val decodedPairIsGood by viewModel.decodedPairIsGood.collectAsState()
        val decodedPairDescs by viewModel.decodedPairDescs.collectAsState()
        val decodedKalakini by viewModel.decodedKalakini.collectAsState()
        val decodedProfile4D by viewModel.decodedProfile4D.collectAsState()
        val decodedMeaning by viewModel.decodedMeaning.collectAsState()
        val pinnedWishMeaning by viewModel.pinnedWishMeaning.collectAsState()
        val hasUserSelectedWishMeaning by viewModel.hasUserSelectedWishMeaning.collectAsState()
        val meaningPool by viewModel.meaningPool.collectAsState()
        val meaningPageIndex by viewModel.meaningPageIndex.collectAsState()
        val isMeaningAutoPlaying by viewModel.isMeaningAutoPlaying.collectAsState()
        val isDecodedFromDb by viewModel.isDecodedFromDb.collectAsState()
        val isDecodingName by viewModel.isDecodingName.collectAsState()
        val decodeError by viewModel.decodeError.collectAsState()
        val randomMeaningSuggestions by viewModel.randomMeaningSuggestions.collectAsState()
        val isAutoSelectingMeaning by viewModel.isAutoSelectingMeaning.collectAsState()
        val uiState by viewModel.uiState.collectAsState()
        val hasCompletedSearch by viewModel.hasCompletedSearch.collectAsState()
        val articles by viewModel.articles.collectAsState()
        val uriHandler = LocalUriHandler.current
        val focusManager = LocalFocusManager.current

        val scrollState = rememberScrollState()
        val birthdayWarningTransition = rememberInfiniteTransition(label = "birthday_warning")
        val birthdayWarningGlow by birthdayWarningTransition.animateFloat(
            initialValue = 0.4f,
            targetValue = 1.0f,
            animationSpec = infiniteRepeatable(
                animation = tween(durationMillis = 800, easing = LinearOutSlowInEasing),
                repeatMode = RepeatMode.Reverse
            ),
            label = "glow_alpha"
        )
        val isBirthdayMissing = dayOfBirth.isEmpty()
        val step1Done = (searchMode == "standalone" && prototypeName.isNotBlank()) || (searchMode == "combined" && surname.isNotBlank())
        val step2Done = step1Done && dayOfBirth.isNotEmpty()
        val step3Done = step2Done && hasUserSelectedWishMeaning && pinnedWishMeaning.isNotBlank()
        // Step 4 is complete only after a successful search, and remains complete
        // when the user returns from the results screen to review the form.
        val step4Done = step3Done && hasCompletedSearch
        val currentStep = when {
            !step1Done -> 1
            !step2Done -> 2
            !step3Done -> 3
            else -> 4
        }
        var showResetConfirmDialog by remember { mutableStateOf(false) }
        val coroutineScope = rememberCoroutineScope()
        var step2Y by remember { mutableStateOf(0) }
        var step3Y by remember { mutableStateOf(0) }
        var step4Y by remember { mutableStateOf(0) }
        var previousSearchMode by remember { mutableStateOf(searchMode) }
        val listState = androidx.compose.foundation.lazy.rememberLazyListState()
        var isAutoScrolling by remember { mutableStateOf(true) }

        // Pause auto-scroll when user drags or touches
        LaunchedEffect(listState.isScrollInProgress) {
            if (listState.isScrollInProgress) {
                isAutoScrolling = false
            }
        }

        // Keep the mode switch easy to follow: after the user changes between
        // standalone and combined mode, return to the beginning of the form.
        // Do not scroll on the initial composition when the screen is first shown.
        LaunchedEffect(searchMode) {
            if (previousSearchMode != searchMode) {
                scrollState.animateScrollTo(0)
                previousSearchMode = searchMode
            }
        }

        // Marquee Effect for Celebrities (using dispatchRawDelta so taps aren't blocked)
        LaunchedEffect(celebrities, isAutoScrolling) {
            if (celebrities.isNotEmpty() && isAutoScrolling) {
                while (isAutoScrolling) {
                    withFrameNanos {
                        listState.dispatchRawDelta(1.2f)
                    }
                }
            }
        }

        var isHeaderVisible by remember { mutableStateOf(true) }

        val nestedScrollConnection = remember {
            object : NestedScrollConnection {
                override fun onPreScroll(available: Offset, source: NestedScrollSource): Offset {
                    if (available.y < -15f && isHeaderVisible && scrollState.value > 10) {
                        isHeaderVisible = false
                    } else if (available.y > 15f && !isHeaderVisible) {
                        isHeaderVisible = true
                    }
                    return Offset.Zero
                }
            }
        }

        LaunchedEffect(scrollState.value) {
            if (scrollState.value == 0) {
                isHeaderVisible = true
            }
        }

        Box(
            modifier = Modifier
                .fillMaxSize()
                .nestedScroll(nestedScrollConnection)
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .background(MaterialTheme.colorScheme.background)
            ) {
                // Unified Collapsible App Bar Container
                Surface(
                    color = GreenHeader,
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .statusBarsPadding()
                    ) {
                        // Collapsible Top Header Row (Logo, Title, Action Buttons)
                        AnimatedVisibility(
                            visible = isHeaderVisible,
                            enter = expandVertically(animationSpec = tween(250)) + fadeIn(animationSpec = tween(200)),
                            exit = shrinkVertically(animationSpec = tween(250)) + fadeOut(animationSpec = tween(200))
                        ) {
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(start = 12.dp, end = 6.dp, top = 6.dp, bottom = 6.dp),
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Image(
                                    painter = painterResource(id = R.drawable.app_logo),
                                    contentDescription = "โลโก้ชื่อดี",
                                    modifier = Modifier
                                        .size(36.dp)
                                        .clip(CircleShape)
                                        .border(1.dp, GoldPrimary.copy(alpha = 0.6f), CircleShape)
                                )
                                Spacer(modifier = Modifier.width(10.dp))
                                Column(modifier = Modifier.weight(1f)) {
                                    Text(
                                        text = "ชื่อดี 4 มิติ",
                                        color = Color.White,
                                        fontWeight = FontWeight.Bold,
                                        fontSize = 16.5.sp,
                                        maxLines = 1
                                    )
                                    Text(
                                        text = "ศาสตร์ตั้งชื่อมงคลอัจฉริยะ",
                                        color = GoldPrimary.copy(alpha = 0.85f),
                                        fontSize = 11.sp,
                                        maxLines = 1
                                    )
                                }
                                IconButton(onClick = onOpenAuspicious) {
                                    Icon(
                                        imageVector = Icons.Default.CalendarMonth,
                                        contentDescription = "ดูฤกษ์ยาม",
                                        tint = GoldPrimary
                                    )
                                }
                                IconButton(onClick = onOpenChat) {
                                    BadgedBox(
                                        badge = {
                                            if (unreadCount > 0) {
                                                Badge(
                                                    containerColor = Color(0xFFD32F2F),
                                                    contentColor = Color.White
                                                ) {
                                                    Text(
                                                        text = if (unreadCount > 99) "99+" else "$unreadCount",
                                                        fontWeight = FontWeight.Bold
                                                    )
                                                }
                                            }
                                        }
                                    ) {
                                        Icon(
                                            imageVector = Icons.AutoMirrored.Filled.Chat,
                                            contentDescription = "ปรึกษาคุณทญา",
                                            tint = GoldPrimary
                                        )
                                    }
                                }
                                IconButton(onClick = { viewModel.toggleGuideSheet(true) }) {
                                    Icon(
                                        imageVector = Icons.AutoMirrored.Filled.LibraryBooks,
                                        contentDescription = "แนะนำ อ.ทญา",
                                        tint = GoldPrimary
                                    )
                                }
                                onOpenDebugPurchase?.let { openDebug ->
                                    IconButton(onClick = openDebug) {
                                        Icon(
                                            imageVector = Icons.Default.BugReport,
                                            contentDescription = "เมนู Debug purchase",
                                            tint = Color(0xFFFF9800)
                                        )
                                    }
                                }
                            }
                        }

                        // Seamless Step Progress Bar (Pinned as lower section of App Bar)
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 4.dp, vertical = 3.dp)
                        ) {
                            WizardStepBar(
                                currentStep = currentStep,
                                step1Done = step1Done,
                                step2Done = step2Done,
                                step3Done = step3Done,
                                step4Done = step4Done,
                                onReset = { showResetConfirmDialog = true },
                                onStepClick = { step ->
                                    coroutineScope.launch {
                                        when (step) {
                                            1 -> scrollState.animateScrollTo(0)
                                            2 -> if (step2Y > 0) scrollState.animateScrollTo((step2Y - 8).coerceAtLeast(0))
                                            3 -> if (step3Y > 0) scrollState.animateScrollTo((step3Y - 8).coerceAtLeast(0))
                                            4 -> if (step4Y > 0) scrollState.animateScrollTo((step4Y - 8).coerceAtLeast(0))
                                        }
                                    }
                                }
                            )
                        }

                        HorizontalDivider(
                            color = GoldPrimary.copy(alpha = 0.35f),
                            thickness = 0.75.dp
                        )
                    }
                }

                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .verticalScroll(scrollState)
                        // One shared page gutter keeps every content frame aligned.
                        .padding(horizontal = 2.dp, vertical = 10.dp),
                    verticalArrangement = Arrangement.spacedBy(10.dp)
                ) {

            // Step 1 is a single bounded section: recommendations, input, and
            // analysis belong together instead of blending into later steps.
            Surface(
                shape = RoundedCornerShape(20.dp),
                color = Color(0xFF160C08),
                border = BorderStroke(1.dp, GoldPrimary.copy(alpha = 0.30f)),
                modifier = Modifier.fillMaxWidth()
            ) {
                Column(
                    modifier = Modifier.padding(bottom = 8.dp),
                    verticalArrangement = Arrangement.spacedBy(10.dp)
                ) {
                    WizardStepHeading(
                        number = "1",
                        title = if (searchMode == "combined") "ระบุนามสกุล (เพื่อหาชื่อคู่มงคล)" else "ระบุชื่อต้นแบบที่คุณชอบ",
                        subtitle = if (searchMode == "combined") "ถอดรหัสคู่เลขและค้นหาชื่อที่เข้ากับนามสกุล" else "ถอดรหัสคู่เลขศาสตร์และค้นหาชื่อใกล้เคียง",
                        isCompleted = step1Done,
                        isActive = currentStep == 1,
                        integratedWithSection = true
                    )

            // ----------------------------------------------------
            // Celebrity Avatars (Famous Names) - Marquee
            // ----------------------------------------------------
            if (searchMode == "standalone") {
                if (celebrities.isNotEmpty()) {
                    val avatarColors = listOf(
                    Color(0xFFE57373), Color(0xFF81C784), Color(0xFF64B5F6), Color(0xFFFFB74D), Color(0xFFBA68C8)
                )

                androidx.compose.foundation.lazy.LazyRow(
                    state = listState,
                    modifier = Modifier.fillMaxWidth(),
                    contentPadding = PaddingValues(horizontal = 16.dp),
                    horizontalArrangement = Arrangement.spacedBy(0.dp)
                ) {
                    items(Int.MAX_VALUE) { i ->
                        val index = i % celebrities.size
                        val celeb = celebrities[index]
                        val name = celeb.name
                        val isSelected = prototypeName == name
                        val initial = name.first().toString()
                        val bgColor = avatarColors[index % avatarColors.size]

                        Column(
                            horizontalAlignment = Alignment.CenterHorizontally,
                            modifier = Modifier
                                .width(80.dp)
                                .padding(vertical = 4.dp)
                                .clickable {
                                    isAutoScrolling = false
                                    viewModel.updatePrototypeNameInput(name)
                                }
                        ) {
                            Box(
                                modifier = Modifier
                                    .size(60.dp)
                                    .then(
                                        if (isSelected) {
                                            Modifier.background(
                                                color = bgColor.copy(alpha = 0.15f),
                                                shape = CircleShape
                                            ).border(
                                                width = 3.5.dp,
                                                color = GoldPrimary,
                                                shape = CircleShape
                                            )
                                        } else {
                                            Modifier.background(
                                                color = bgColor.copy(alpha = 0.15f),
                                                shape = CircleShape
                                            ).border(
                                                width = 2.5.dp,
                                                color = bgColor,
                                                shape = CircleShape
                                            )
                                        }
                                    ),
                                contentAlignment = Alignment.Center
                            ) {
                                if (!celeb.avatar_url.isNullOrEmpty()) {
                                    AsyncImage(
                                        model = MiracleApiConfig.assetUrl("avatars/${celeb.avatar_url}"),
                                        contentDescription = name,
                                        contentScale = ContentScale.Crop,
                                        modifier = Modifier
                                            .size(60.dp)
                                            .clip(CircleShape)
                                    )
                                } else {
                                    Text(
                                        text = initial,
                                        color = Color.White,
                                        fontSize = 14.sp,
                                        fontWeight = FontWeight.Bold
                                    )
                                }
                            }
                            Spacer(modifier = Modifier.height(6.dp))
                            Text(
                                text = name.split(" ")[0],
                                color = if (isSelected) GoldPrimary else GoldLight.copy(alpha = 0.9f),
                                fontSize = 14.sp,
                                fontWeight = FontWeight.W700,
                                maxLines = 1,
                                textAlign = androidx.compose.ui.text.style.TextAlign.Center,
                                overflow = androidx.compose.ui.text.style.TextOverflow.Ellipsis
                            )
                        }
                    }
                }
            }
            }

            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color(0xFF160C08), RoundedCornerShape(16.dp))
                    .padding(10.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                // Mode Selector (Top - determines input field mode)
                ModeSegmentedControl(
                    selectedMode = searchMode,
                    onModeSelected = viewModel::updateSearchMode
                )

                // Prototype Name input: AI Wish Query Text & Analysis
                AnimatedVisibility(
                    visible = searchMode == "standalone",
                    enter = androidx.compose.animation.expandVertically() + androidx.compose.animation.fadeIn(),
                    exit = androidx.compose.animation.shrinkVertically() + androidx.compose.animation.fadeOut()
                ) {
                    Column(
                        modifier = Modifier.fillMaxWidth(),
                        verticalArrangement = Arrangement.spacedBy(0.dp)
                    ) {
                        OutlinedTextField(
                            value = prototypeName,
                            onValueChange = {
                                viewModel.updatePrototypeNameInput(it)
                            },
                            label = { Text("พิมพ์ชื่อเพื่อค้นหาชื่อใกล้เคียง") },
                            leadingIcon = {
                                Icon(
                                    imageVector = Icons.Default.EditNote,
                                    contentDescription = null
                                )
                            },
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(14.dp),
                            singleLine = true,
                            isError = decodeError != null,
                            supportingText = decodeError?.let { message ->
                                { Text(text = message, fontSize = 11.sp) }
                            },
                            colors = OutlinedTextFieldDefaults.colors(
                                focusedBorderColor = GoldPrimary,
                                unfocusedBorderColor = GoldPrimary.copy(alpha = 0.55f),
                                focusedLabelColor = GoldLight,
                                unfocusedLabelColor = GoldLight.copy(alpha = 0.78f),
                                focusedLeadingIconColor = GoldPrimary,
                                unfocusedLeadingIconColor = GoldLight.copy(alpha = 0.72f),
                                focusedTrailingIconColor = GoldLight,
                                unfocusedTrailingIconColor = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.72f),
                                focusedContainerColor = Color(0xFF2C1A11),
                                unfocusedContainerColor = Color(0xFF26160E),
                                cursorColor = GoldPrimary,
                                focusedTextColor = MaterialTheme.colorScheme.onSurface,
                                unfocusedTextColor = MaterialTheme.colorScheme.onSurface
                            ),
                            trailingIcon = {
                                if (isDecodingName) {
                                    CircularProgressIndicator(
                                        modifier = Modifier.size(18.dp),
                                        strokeWidth = 2.dp,
                                        color = GoldPrimary
                                    )
                                } else if (prototypeName.isNotEmpty()) {
                                    IconButton(onClick = {
                                        viewModel.updatePrototypeNameInput("")
                                    }) {
                                        Icon(imageVector = Icons.Default.Clear, contentDescription = "ล้างข้อความ")
                                    }
                                }
                            }
                        )

                        // ----------------------------------------------------
                        // Real-time Analysis Section (Backend API decoded pairs)
                        // ----------------------------------------------------
                        AnimatedVisibility(
                            visible = prototypeName.isNotBlank() && decodedPairs.isNotEmpty(),
                            enter = androidx.compose.animation.expandVertically() + androidx.compose.animation.fadeIn(),
                            exit = androidx.compose.animation.shrinkVertically() + androidx.compose.animation.fadeOut()
                        ) {
                            Surface(
                                shape = RoundedCornerShape(
                                    topStart = 4.dp,
                                    topEnd = 4.dp,
                                    bottomStart = 12.dp,
                                    bottomEnd = 12.dp
                                ),
                                color = Color(0xFF1F120B).copy(alpha = 0.92f),
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(top = 2.dp)
                            ) {
                                Column(
                                    modifier = Modifier.padding(horizontal = 8.dp, vertical = 10.dp),
                                    verticalArrangement = Arrangement.spacedBy(8.dp)
                                ) {
                                    if (isDecodedFromDb && decodedMeaning.isNotBlank()) {
                                        Text(
                                            text = "💡 ความหมายของ “${prototypeName.trim()}”",
                                            fontSize = 12.sp,
                                            fontWeight = FontWeight.Bold,
                                            color = GoldLight
                                        )
                                        val formattedMeaning = decodedMeaning.trim()
                                        val displayMeaning = if ((formattedMeaning.startsWith("“") && formattedMeaning.endsWith("”")) ||
                                            (formattedMeaning.startsWith("\"") && formattedMeaning.endsWith("\""))
                                        ) {
                                            formattedMeaning
                                        } else {
                                            "“${formattedMeaning.trim('"', '“', '”', '\'')}”"
                                        }
                                        Text(
                                            text = displayMeaning,
                                            fontSize = 13.sp,
                                            color = Color.White.copy(alpha = 0.92f),
                                            lineHeight = 18.sp,
                                            textAlign = androidx.compose.ui.text.style.TextAlign.Center,
                                            modifier = Modifier.fillMaxWidth()
                                        )
                                        HorizontalDivider(
                                            modifier = Modifier.padding(vertical = 2.dp),
                                            color = GoldPrimary.copy(alpha = 0.35f),
                                            thickness = 0.5.dp
                                        )
                                    }

                                    Surface(
                                        shape = RoundedCornerShape(10.dp),
                                        color = Color(0xFF1F120B).copy(alpha = 0.95f),
                                        modifier = Modifier.fillMaxWidth()
                                    ) {
                                        Column(
                                            modifier = Modifier.padding(horizontal = 12.dp, vertical = 12.dp),
                                            verticalArrangement = Arrangement.spacedBy(10.dp)
                                        ) {
                                            Box(
                                                modifier = Modifier.fillMaxWidth(),
                                                contentAlignment = Alignment.Center
                                            ) {
                                                Surface(
                                                    shape = RoundedCornerShape(6.dp),
                                                    color = GoldPrimary.copy(alpha = 0.18f),
                                                    border = BorderStroke(1.dp, GoldPrimary.copy(alpha = 0.6f))
                                                ) {
                                                    Text(
                                                        text = "“${prototypeName.trim()}”",
                                                        fontSize = 15.sp,
                                                        fontWeight = FontWeight.Bold,
                                                        color = GoldLight,
                                                        modifier = Modifier.padding(horizontal = 10.dp, vertical = 3.dp)
                                                    )
                                                }
                                            }

                                            Box(
                                                modifier = Modifier.fillMaxWidth(),
                                                contentAlignment = Alignment.Center
                                            ) {
                                                PairCircles(
                                                    pairsStr = decodedPairs.joinToString(","),
                                                    typesStr = decodedPairTypes.joinToString(","),
                                                    goodFlags = decodedPairIsGood,
                                                    onPairClick = { pair -> viewModel.fetchPairDetail(pair) }
                                                )
                                            }

                                            if (decodedKalakini.applied) {
                                                val hasMatches = decodedKalakini.matches.isNotBlank()
                                                val formattedMatches = decodedKalakini.matches.filter { !it.isWhitespace() }.toList().joinToString(" ")
                                                Text(
                                                    text = if (hasMatches) {
                                                        "พบกาลกิณีในชื่อ: $formattedMatches"
                                                    } else {
                                                        "ไม่พบกาลกิณีในชื่อ"
                                                    },
                                                    fontSize = 11.sp,
                                                    textAlign = androidx.compose.ui.text.style.TextAlign.Center,
                                                    modifier = Modifier.fillMaxWidth(),
                                                    color = if (hasMatches) Color(0xFFFFB74D) else Color(0xFF81C784)
                                                )
                                            }

                                            if (decodedProfile4D != null) {
                                                NameProfile4DCard(
                                                    profile = decodedProfile4D!!,
                                                    onWeightsChanged = { n, sem, snd, sim ->
                                                        viewModel.updateRankingWeights(n, sem, snd, sim)
                                                    },
                                                    modifier = Modifier.padding(top = 4.dp)
                                                )
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }

                // If combined mode, show Surname field
                AnimatedVisibility(
                    visible = searchMode == "combined",
                    enter = androidx.compose.animation.expandVertically() + androidx.compose.animation.fadeIn(),
                    exit = androidx.compose.animation.shrinkVertically() + androidx.compose.animation.fadeOut()
                ) {
                    Column(
                        modifier = Modifier.fillMaxWidth(),
                        verticalArrangement = Arrangement.spacedBy(0.dp)
                    ) {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(end = 2.dp, bottom = 4.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.End
                        ) {
                            Surface(
                                modifier = Modifier
                                    .height(30.dp)
                                    .clickable(onClick = viewModel::toggleCombinedSearchTarget),
                                shape = RoundedCornerShape(999.dp),
                                color = GoldPrimary.copy(alpha = 0.12f),
                                border = BorderStroke(0.8.dp, GoldPrimary.copy(alpha = 0.38f))
                            ) {
                                Row(
                                    modifier = Modifier.padding(horizontal = 12.dp),
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.spacedBy(6.dp)
                                ) {
                                    Text(
                                        text = if (isFindSurnameMode) "กำลังหา: นามสกุล แตะเพื่อสลับหาชื่อตัว"
                                               else "กำลังหา: ชื่อตัว แตะเพื่อสลับหาสกุล",
                                        fontSize = 10.5.sp,
                                        fontWeight = FontWeight.SemiBold,
                                        color = GoldLight,
                                        maxLines = 1
                                    )
                                    Text(
                                        text = "⇄",
                                        fontSize = 11.sp,
                                        fontWeight = FontWeight.Bold,
                                        color = GoldLight
                                    )
                                }
                            }
                        }
                        OutlinedTextField(
                            value = surname,
                            onValueChange = {
                                viewModel.surname.value = it
                                viewModel.decodeCurrentName(it)
                            },
                            label = {
                                Text(
                                    if (isFindSurnameMode) "พิมพ์ชื่อตัวที่ต้องการใช้ร่วมกัน"
                                    else "พิมพ์นามสกุลที่ต้องการใช้ร่วมกัน"
                                )
                            },
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(14.dp),
                            singleLine = true,
                            isError = decodeError != null,
                            supportingText = decodeError?.let { message ->
                                { Text(text = message, fontSize = 11.sp) }
                            },
                            colors = OutlinedTextFieldDefaults.colors(
                                focusedBorderColor = GoldPrimary,
                                unfocusedBorderColor = Color.Gray,
                                focusedTextColor = MaterialTheme.colorScheme.onSurface,
                                unfocusedTextColor = MaterialTheme.colorScheme.onSurface
                            ),
                            trailingIcon = {
                                if (isDecodingName) {
                                    CircularProgressIndicator(
                                        modifier = Modifier.size(18.dp),
                                        strokeWidth = 2.dp,
                                        color = GoldPrimary
                                    )
                                } else if (surname.isNotEmpty()) {
                                    IconButton(onClick = {
                                        viewModel.surname.value = ""
                                        viewModel.decodeCurrentName("")
                                    }) {
                                        Icon(imageVector = Icons.Default.Clear, contentDescription = "ล้างข้อความ")
                                    }
                                }
                            }
                        )

                        AnimatedVisibility(
                            visible = surname.isNotBlank() && decodedPairs.isNotEmpty(),
                            enter = androidx.compose.animation.expandVertically() + androidx.compose.animation.fadeIn(),
                            exit = androidx.compose.animation.shrinkVertically() + androidx.compose.animation.fadeOut()
                        ) {
                            Surface(
                                shape = RoundedCornerShape(
                                    topStart = 4.dp,
                                    topEnd = 4.dp,
                                    bottomStart = 12.dp,
                                    bottomEnd = 12.dp
                                ),
                                color = Color(0xFF1F120B),
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(top = 2.dp)
                            ) {
                                Column(
                                    modifier = Modifier.padding(horizontal = 8.dp, vertical = 10.dp),
                                    verticalArrangement = Arrangement.spacedBy(8.dp)
                                ) {
                                    if (isDecodedFromDb && decodedMeaning.isNotBlank()) {
                                        Text(
                                            text = if (isFindSurnameMode) {
                                                "💡 ความหมายของชื่อตัว “${surname.trim()}”"
                                            } else {
                                                "💡 ความหมายของนามสกุล “${surname.trim()}”"
                                            },
                                            fontSize = 12.sp,
                                            fontWeight = FontWeight.Bold,
                                            color = GoldLight
                                        )
                                        val formattedSurnameMeaning = decodedMeaning.trim()
                                        val displaySurnameMeaning = if ((formattedSurnameMeaning.startsWith("“") && formattedSurnameMeaning.endsWith("”")) ||
                                            (formattedSurnameMeaning.startsWith("\"") && formattedSurnameMeaning.endsWith("\""))
                                        ) {
                                            formattedSurnameMeaning
                                        } else {
                                            "“${formattedSurnameMeaning.trim('"', '“', '”', '\'')}”"
                                        }
                                        Text(
                                            text = displaySurnameMeaning,
                                            fontSize = 13.sp,
                                            color = Color.White.copy(alpha = 0.92f),
                                            lineHeight = 18.sp,
                                            textAlign = androidx.compose.ui.text.style.TextAlign.Center,
                                            modifier = Modifier.fillMaxWidth()
                                        )
                                        HorizontalDivider(
                                            modifier = Modifier.padding(vertical = 2.dp),
                                            color = GoldPrimary.copy(alpha = 0.35f),
                                            thickness = 0.5.dp
                                        )
                                    }

                                    Surface(
                                        shape = RoundedCornerShape(10.dp),
                                        color = Color(0xFF1F120B).copy(alpha = 0.95f),
                                        modifier = Modifier.fillMaxWidth()
                                    ) {
                                        Column(
                                            modifier = Modifier.padding(horizontal = 12.dp, vertical = 12.dp),
                                            verticalArrangement = Arrangement.spacedBy(10.dp)
                                        ) {
                                            Row(
                                                modifier = Modifier.fillMaxWidth(),
                                                horizontalArrangement = Arrangement.Center,
                                                verticalAlignment = Alignment.CenterVertically
                                            ) {
                                                Text(
                                                    text = "คู่เลขศาสตร์ และพลังเงาของนามสกุล",
                                                    fontSize = 12.5.sp,
                                                    fontWeight = FontWeight.Medium,
                                                    color = Color.White.copy(alpha = 0.85f)
                                                )
                                            }

                                            Box(
                                                modifier = Modifier.fillMaxWidth(),
                                                contentAlignment = Alignment.Center
                                            ) {
                                                PairCircles(
                                                    pairsStr = decodedPairs.joinToString(","),
                                                    typesStr = decodedPairTypes.joinToString(","),
                                                    goodFlags = decodedPairIsGood,
                                                    onPairClick = { pair -> viewModel.fetchPairDetail(pair) }
                                                )
                                            }

                                            if (decodedProfile4D != null) {
                                                NameProfile4DCard(
                                                    profile = decodedProfile4D!!,
                                                    onWeightsChanged = { n, sem, snd, sim ->
                                                        viewModel.updateRankingWeights(n, sem, snd, sim)
                                                    },
                                                    modifier = Modifier.padding(top = 4.dp)
                                                )
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }

            }
                }
            }

            // ====================================================
            // STEP 2: ระบุวันเกิดประจำตัว
            // ====================================================
            Surface(
                shape = RoundedCornerShape(20.dp),
                color = Color(0xFF160C08),
                border = BorderStroke(1.dp, GoldPrimary.copy(alpha = 0.30f)),
                modifier = Modifier
                    .fillMaxWidth()
                    .onGloballyPositioned { coordinates ->
                        val y = coordinates.positionInParent().y.toInt()
                        if (y > 0) step2Y = y
                    }
            ) {
                Column {
            WizardStepHeading(
                number = "2",
                title = "ระบุวันเกิดประจำตัว",
                subtitle = "คัดกรองอักษรกาลกิณีให้ถูกต้องตามวันเกิด",
                isCompleted = step2Done,
                isActive = currentStep == 2,
                integratedWithSection = true
            )

            Surface(
                shape = RoundedCornerShape(0.dp),
                color = Color(0xFF160C08),
                modifier = Modifier.fillMaxWidth()
            ) {
                Column {
                    val daysList = listOf(
                        "sunday" to "วันอาทิตย์",
                        "monday" to "วันจันทร์",
                        "tuesday" to "วันอังคาร",
                        "wednesday1" to "วันพุธ (กลางวัน)",
                        "wednesday2" to "วันพุธ (กลางคืน)",
                        "thursday" to "วันพฤหัสบดี",
                        "friday" to "วันศุกร์",
                        "saturday" to "วันเสาร์"
                    )

                    var dayDropdownExpanded by remember { mutableStateOf(false) }
                    val selectedDayName = daysList.find { it.first == dayOfBirth }?.second ?: "แตะเพื่อเลือกวันเกิด"

                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 10.dp, vertical = 9.dp)
                    ) {
                        Box(modifier = Modifier.fillMaxWidth()) {
                            Surface(
                                onClick = { dayDropdownExpanded = true },
                                modifier = Modifier.fillMaxWidth(),
                                shape = RoundedCornerShape(12.dp),
                                color = Color(0xFF26160E)
                            ) {
                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(horizontal = 14.dp, vertical = 11.dp),
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    Icon(
                                        Icons.Default.CalendarMonth,
                                        contentDescription = null,
                                        tint = GoldPrimary,
                                        modifier = Modifier.size(28.dp)
                                    )
                                    Spacer(Modifier.width(13.dp))
                                    Column(modifier = Modifier.weight(1f)) {
                                        Text(
                                            "วันเกิดประจำตัว",
                                            fontSize = 12.sp,
                                            fontWeight = FontWeight.Medium,
                                            color = GoldLight.copy(alpha = 0.72f)
                                        )
                                        Text(
                                            selectedDayName.removeSuffix("..."),
                                            fontSize = 16.sp,
                                            fontWeight = FontWeight.SemiBold,
                                            color = if (dayOfBirth.isEmpty()) GoldLight.copy(alpha = 0.58f)
                                            else MaterialTheme.colorScheme.onSurface
                                        )
                                    }
                                    ExposedDropdownMenuDefaults.TrailingIcon(
                                        expanded = dayDropdownExpanded
                                    )
                                }
                            }
                            DropdownMenu(
                                expanded = dayDropdownExpanded,
                                onDismissRequest = { dayDropdownExpanded = false }
                            ) {
                                daysList.forEach { (key, name) ->
                                    DropdownMenuItem(
                                        text = { Text(name, fontWeight = if (dayOfBirth == key) FontWeight.Bold else FontWeight.Normal, color = if (dayOfBirth == key) GoldPrimary else MaterialTheme.colorScheme.onSurface) },
                                        onClick = {
                                            viewModel.dayOfBirth.value = key
                                            if (prototypeName.isNotBlank()) {
                                                viewModel.refreshSelectedNameForBirthday()
                                            }
                                            dayDropdownExpanded = false
                                        }
                                    )
                                }
                            }
                        }
                    }

                    if (dayOfBirth.isNotBlank()) {
                        val dayInfo = getKalakiniDayInfo(dayOfBirth)
                        if (dayInfo != null) {
                            KalakiniSummaryCard(
                                dayInfo = dayInfo
                            )
                        }
                    }
                }
            }
                }
            }

            // ====================================================
            // STEP 3: เสริมความหมายมงคลที่ต้องการ
            // ====================================================
            Surface(
                shape = RoundedCornerShape(20.dp),
                color = Color(0xFF160C08),
                border = BorderStroke(1.dp, GoldPrimary.copy(alpha = 0.30f)),
                modifier = Modifier
                    .fillMaxWidth()
                    .onGloballyPositioned { coordinates ->
                        val y = coordinates.positionInParent().y.toInt()
                        if (y > 0) step3Y = y
                    }
            ) {
                Column {
            WizardStepHeading(
                number = "3",
                title = "เสริมความหมายมงคลที่ต้องการ",
                subtitle = "เลือกความหมายที่ต้องการหนุนนำ หรือสุ่มใหม่",
                isCompleted = step3Done,
                isActive = currentStep == 3,
                integratedWithSection = true
            )

            var currentWishList by remember { mutableStateOf(emptyList<String>()) }

            // Auto-play timer: advances meaning page every 8 seconds when playing
            LaunchedEffect(Unit) {
                if (meaningPool.isEmpty()) {
                    viewModel.loadRandomMeanings(pinFirst = false)
                }
            }

            LaunchedEffect(isMeaningAutoPlaying) {
                while (isMeaningAutoPlaying) {
                    delay(5000L)
                    viewModel.nextMeaningPage()
                }
            }

            LaunchedEffect(hasUserSelectedWishMeaning, pinnedWishMeaning, randomMeaningSuggestions) {
                val base = if (randomMeaningSuggestions.isNotEmpty()) {
                    randomMeaningSuggestions
                } else {
                    NameSearchViewModel.DEFAULT_AUSPICIOUS_MEANINGS.shuffled().take(3)
                }
                currentWishList = base
                    .map { if (it.startsWith("✨")) it else "✨ $it" }
            }

            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 10.dp, vertical = 9.dp),
                colors = CardDefaults.cardColors(containerColor = Color(0xFF1F120B)),
                shape = RoundedCornerShape(14.dp)
            ) {
                Column(modifier = Modifier.fillMaxWidth()) {
                    val hasSelection = hasUserSelectedWishMeaning && pinnedWishMeaning.isNotBlank()
                    if (decodedMeaning.isNotBlank()) {
                        val cleanProto = decodedMeaning.trim('"', '“', '”', '\'', ' ')
                        val displayProto = "“$cleanProto”"
                        val isProtoSelected = hasSelection && (pinnedWishMeaning == cleanProto || pinnedWishMeaning == decodedMeaning)
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .clickable {
                                    viewModel.pinWishMeaning(decodedMeaning)
                                }
                                .background(
                                    if (isProtoSelected) {
                                        Brush.horizontalGradient(listOf(Color(0xFF522E1A), Color(0xFF381F12), Color(0xFF26150C)))
                                    } else {
                                        Brush.horizontalGradient(listOf(Color(0xFF26160E), Color(0xFF1F120B)))
                                    }
                                )
                                .defaultMinSize(minHeight = 52.dp)
                                .padding(
                                    start = if (isProtoSelected) 0.dp else 14.dp,
                                    end = 14.dp,
                                    top = 10.dp,
                                    bottom = 10.dp
                                ),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            if (isProtoSelected) {
                                Box(
                                    modifier = Modifier
                                        .width(4.dp)
                                        .height(36.dp)
                                        .background(
                                            brush = Brush.verticalGradient(
                                                listOf(Color(0xFFFFDF85), GoldPrimary)
                                            ),
                                            shape = RoundedCornerShape(topEnd = 2.dp, bottomEnd = 2.dp)
                                        )
                                )
                                Spacer(Modifier.width(10.dp))
                            }
                            Text(if (isProtoSelected) "✅" else "🏷️", fontSize = 16.sp)
                            Spacer(Modifier.width(10.dp))
                            Column(modifier = Modifier.weight(1f)) {
                                Text(
                                    text = "ความหมายจากชื่อต้นแบบ ${if (prototypeName.isNotBlank()) "($prototypeName)" else ""}",
                                    fontSize = 11.sp,
                                    fontWeight = FontWeight.Bold,
                                    color = if (isProtoSelected) Color(0xFFFFDF85) else GoldPrimary
                                )
                                Spacer(Modifier.height(2.dp))
                                Text(
                                    text = displayProto,
                                    fontSize = if (isProtoSelected) 14.5.sp else 13.5.sp,
                                    fontWeight = if (isProtoSelected) FontWeight.Bold else FontWeight.Normal,
                                    color = if (isProtoSelected) Color(0xFFFFDF85) else Color.White,
                                    maxLines = 2,
                                    overflow = androidx.compose.ui.text.style.TextOverflow.Ellipsis,
                                    lineHeight = if (isProtoSelected) 20.sp else 18.sp
                                )
                            }
                        }
                    }
                    if (isAutoSelectingMeaning && currentWishList.isEmpty()) {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 14.dp, vertical = 14.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(10.dp)
                        ) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(16.dp),
                                strokeWidth = 2.dp,
                                color = GoldPrimary
                            )
                            Text(
                                text = "กำลังสุ่มความหมายมงคลให้คุณ...",
                                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.78f),
                                fontSize = 11.sp
                            )
                        }
                    }
                    currentWishList.forEachIndexed { index, targetWish ->
                        AnimatedContent(
                            targetState = targetWish,
                            transitionSpec = {
                                val duration = 420
                                (slideInVertically(
                                    animationSpec = tween(duration, easing = FastOutSlowInEasing),
                                    initialOffsetY = { height -> height }
                                ) + fadeIn(animationSpec = tween(duration, easing = FastOutSlowInEasing))) togetherWith
                                (slideOutVertically(
                                    animationSpec = tween(duration, easing = FastOutSlowInEasing),
                                    targetOffsetY = { height -> -height }
                                ) + fadeOut(animationSpec = tween(duration, easing = FastOutSlowInEasing)))
                            },
                            label = "wishRow$index"
                        ) { animatedWish ->
                            key(animatedWish) {
                            val wish = animatedWish

                            // Flash + scale bounce when newly selected (non-pinned)
                            var triggerFlash by remember { mutableStateOf(false) }
                            val flashAlpha by animateFloatAsState(
                                targetValue = if (triggerFlash) 0f else 1f,
                                animationSpec = tween(durationMillis = 320),
                                label = "flashAlpha",
                                finishedListener = { triggerFlash = false }
                            )
                            val selectScale by animateFloatAsState(
                                targetValue = if (triggerFlash) 1.035f else 1f,
                                animationSpec = tween(durationMillis = 200, easing = FastOutSlowInEasing),
                                label = "selectScale"
                            )

                            val cleanWish = wish.removePrefix("✨").trim('"', '“', '”', '\'', ' ')
                            val displayWish = "“$cleanWish”"
                            val selected = hasSelection && (pinnedWishMeaning == cleanWish || pinnedWishMeaning == wish.removePrefix("✨").trim())

                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .graphicsLayer {
                                        scaleX = selectScale
                                        scaleY = selectScale
                                    }
                            ) {
                                val itemShape = RoundedCornerShape(0.dp)
                                val selectedShape = RoundedCornerShape(0.dp)
                                val selectedBackground = Brush.horizontalGradient(
                                    colors = listOf(
                                        Color(0xFF522E1A),
                                        Color(0xFF381F12),
                                        Color(0xFF26150C)
                                    )
                                )
                                val suggestionBackground = Brush.horizontalGradient(
                                    colors = if (index % 2 == 0) listOf(
                                        Color(0xFF26160E),
                                        Color(0xFF1F120B)
                                    ) else listOf(
                                        Color(0xFF2C1A11),
                                        Color(0xFF22140D)
                                    )
                                )
                                Column(modifier = Modifier.fillMaxWidth()) {
                                    Row(
                                        modifier = Modifier
                                            .fillMaxWidth()
                                            .clip(if (selected) selectedShape else itemShape)
                                            .clickable {
                                                if (!selected) {
                                                    triggerFlash = true
                                                    viewModel.pinWishMeaning(wish)
                                                }
                                            }
                                            .then(
                                                if (selected) {
                                                    Modifier.background(selectedBackground)
                                                } else {
                                                    Modifier.background(suggestionBackground)
                                                }
                                            )
                                            .defaultMinSize(minHeight = 48.dp)
                                            .padding(
                                                start = if (selected) 0.dp else 14.dp,
                                                end = 14.dp,
                                                top = 8.dp,
                                                bottom = 8.dp
                                            ),
                                        verticalAlignment = Alignment.CenterVertically
                                    ) {
                                        if (selected) {
                                            Box(
                                                modifier = Modifier
                                                    .width(4.dp)
                                                    .height(34.dp)
                                                    .background(
                                                        brush = Brush.verticalGradient(
                                                            listOf(Color(0xFFFFDF85), GoldPrimary)
                                                        ),
                                                        shape = RoundedCornerShape(topEnd = 2.dp, bottomEnd = 2.dp)
                                                    )
                                            )
                                            Spacer(Modifier.width(10.dp))
                                        }
                                        Text(
                                            text = if (selected) "✅" else "✨",
                                            fontSize = 15.sp
                                        )
                                        Spacer(Modifier.width(8.dp))
                                        Text(
                                            text = displayWish,
                                            fontSize = if (selected) 14.5.sp else 13.5.sp,
                                            fontWeight = if (selected) FontWeight.Bold else FontWeight.Normal,
                                            color = if (selected) Color(0xFFFFDF85) else MaterialTheme.colorScheme.onSurface,
                                            modifier = Modifier.weight(1f),
                                            maxLines = 2,
                                            overflow = androidx.compose.ui.text.style.TextOverflow.Ellipsis,
                                            lineHeight = if (selected) 20.sp else 18.sp
                                        )
                                    }
                                }
                                // Gold flash overlay — inside BoxScope
                                if (triggerFlash) {
                                    Box(
                                        modifier = Modifier
                                            .matchParentSize()
                                            .clip(itemShape)
                                            .background(GoldPrimary.copy(alpha = 0.28f * flashAlpha))
                                    )
                                }
                            }
                        }
                    }
                    }

                    // Media-style Control Bar (Back, Play/Pause, Forward, Counter)
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(Color(0xFF1F120B))
                            .padding(horizontal = 12.dp, vertical = 8.dp),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(10.dp)
                    ) {
                        // Back button
                        IconButton(
                            onClick = { viewModel.previousMeaningPage() },
                            modifier = Modifier
                                .size(32.dp)
                                .background(Color(0xFF2C1A11), CircleShape)
                                .border(1.dp, GoldPrimary.copy(alpha = 0.35f), CircleShape)
                        ) {
                            Text(
                                text = "⏮",
                                fontSize = 14.sp,
                                color = GoldPrimary
                            )
                        }

                        // Play / Pause button
                        Surface(
                            shape = CircleShape,
                            color = Color(0xFF2C1A11),
                            border = BorderStroke(1.dp, GoldPrimary.copy(alpha = 0.40f)),
                            modifier = Modifier.clickable { viewModel.toggleMeaningAutoPlay() }
                        ) {
                            Row(
                                modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(6.dp)
                            ) {
                                Text(
                                    text = if (isMeaningAutoPlaying) "⏸" else "▶",
                                    fontSize = 11.sp,
                                    color = GoldPrimary
                                )
                                Text(
                                    text = if (isMeaningAutoPlaying) "หยุดเลื่อน" else "เลื่อนต่อ",
                                    fontSize = 11.sp,
                                    fontWeight = FontWeight.SemiBold,
                                    color = GoldPrimary
                                )
                            }
                        }

                        // Forward button
                        IconButton(
                            onClick = { viewModel.nextMeaningPage() },
                            modifier = Modifier
                                .size(32.dp)
                                .background(Color(0xFF2C1A11), CircleShape)
                                .border(1.dp, GoldPrimary.copy(alpha = 0.35f), CircleShape)
                        ) {
                            Text(
                                text = "⏭",
                                fontSize = 14.sp,
                                color = GoldPrimary
                            )
                        }

                        Spacer(Modifier.weight(1f))

                        val totalPages = viewModel.totalMeaningPages()
                        Text(
                            text = "ชุดที่ ${meaningPageIndex + 1}/$totalPages",
                            fontSize = 11.sp,
                            fontWeight = FontWeight.Medium,
                            color = GoldLight.copy(alpha = 0.75f)
                        )
                    }
                }
            }
            }
            }

            // ====================================================
            // STEP 4: ยืนยันโปรไฟล์ชื่อดี 4 มิติ & ค้นหา
            // ====================================================
            Surface(
                shape = RoundedCornerShape(20.dp),
                color = Color(0xFF160C08),
                border = BorderStroke(1.dp, GoldPrimary.copy(alpha = 0.30f)),
                modifier = Modifier
                    .fillMaxWidth()
                    .onGloballyPositioned { coordinates ->
                        val y = coordinates.positionInParent().y.toInt()
                        if (y > 0) step4Y = y
                    }
            ) {
                Column {
            WizardStepHeading(
                number = "4",
                title = "ยืนยันโปรไฟล์ 4 มิติ & ค้นหาชื่อมงคล",
                isCompleted = step4Done,
                isActive = currentStep == 4,
                integratedWithSection = true
            )

            // One ternary control replaces multiple presets and sliders (Magical Design).
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 10.dp, vertical = 9.dp),
                shape = RoundedCornerShape(14.dp),
                colors = CardDefaults.cardColors(containerColor = Color(0xFF1F120B)),
                elevation = CardDefaults.cardElevation(defaultElevation = 0.dp)
            ) {
            Column(modifier = Modifier.padding(14.dp)) {
            Card(
                modifier = Modifier
                    .fillMaxWidth(),
                shape = RoundedCornerShape(12.dp),
                colors = CardDefaults.cardColors(containerColor = Color.Transparent),
                elevation = CardDefaults.cardElevation(defaultElevation = 0.dp)
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(
                            Brush.linearGradient(
                                colors = listOf(
                                    Color(0xFF26160E),
                                    Color(0xFF1F120B),
                                    Color(0xFF160C08)
                                )
                            )
                        )
                        .padding(bottom = 4.dp)
                ) {
                    Column {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.SpaceBetween
                        ) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Text(
                                    text = "⚖️",
                                    fontSize = 16.sp
                                )
                                Spacer(modifier = Modifier.width(6.dp))
                                Text(
                                    text = "ยืนยันโปรไฟล์ชื่อดี 4 มิติ",
                                    color = Color.White,
                                    fontSize = 15.sp,
                                    fontWeight = FontWeight.Bold
                                )
                            }
                            Surface(
                                modifier = Modifier
                                    .clickable(onClick = viewModel::resetRankingWeightsToBalance),
                                shape = RoundedCornerShape(999.dp),
                                color = GoldPrimary.copy(alpha = 0.12f),
                                border = BorderStroke(0.8.dp, GoldPrimary.copy(alpha = 0.38f))
                            ) {
                                Row(
                                    modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.spacedBy(4.dp)
                                ) {
                                    Text(
                                        text = "↺",
                                        fontSize = 12.sp,
                                        fontWeight = FontWeight.Bold,
                                        color = GoldLight
                                    )
                                    Text(
                                        text = "คืนค่าจุดสมดุล",
                                        fontSize = 10.sp,
                                        fontWeight = FontWeight.Medium,
                                        color = GoldLight
                                    )
                                }
                            }
                        }
                        val refName = if (searchMode == "combined") surname.trim() else prototypeName.trim()
                        if (refName.isNotEmpty()) {
                            Text(
                                text = "✨ อ้างอิงจากต้นแบบ “$refName”",
                                fontSize = 11.5.sp,
                                fontWeight = FontWeight.SemiBold,
                                color = GoldPrimary,
                                modifier = Modifier.padding(top = 2.dp)
                            )
                        }
                            Text(
                                text = "แสดงโปรไฟล์เดียวกับผลวิเคราะห์ชื่อที่กรอก",
                            fontSize = 11.sp,
                            color = GoldLight.copy(alpha = 0.85f),
                            modifier = Modifier.padding(top = 1.dp, bottom = 6.dp)
                        )

                        val rawWeights = listOf(
                            pendingWeights.numerology,
                            pendingWeights.semantic,
                            pendingWeights.similarity,
                            pendingWeights.sound
                        )
                        val weightTotal = rawWeights.sum().coerceAtLeast(0.0001f)
                        val scaled = rawWeights.map { (it / weightTotal) * 100f }
                        val intParts = scaled.map { kotlin.math.floor(it).toInt() }.toMutableList()
                        var remainder = 100 - intParts.sum()
                        val sortedIndices = scaled.indices.sortedByDescending { scaled[it] - kotlin.math.floor(scaled[it]) }
                        for (idx in sortedIndices) {
                            if (remainder <= 0) break
                            intParts[idx]++
                            remainder--
                        }
                        val numerologyPercent = intParts[0]
                        val semanticPercent = intParts[1]
                        val similarityPercent = intParts[2]
                        val soundPercent = intParts[3]
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(top = 6.dp),
                            horizontalArrangement = Arrangement.spacedBy(8.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            WizardWeightPercentStack(
                                numerologyPercent = numerologyPercent,
                                semanticPercent = semanticPercent,
                                soundPercent = soundPercent,
                                similarityPercent = similarityPercent,
                                modifier = Modifier.width(125.dp)
                            )
                            RadarWeightSelector(
                                numerologyWeight = pendingWeights.numerology,
                                semanticWeight = pendingWeights.semantic,
                                soundWeight = pendingWeights.sound,
                                similarityWeight = pendingWeights.similarity,
                                onWeightsChanged = { n, sem, snd, sim ->
                                    viewModel.updateRankingWeights(n, sem, snd, sim)
                                },
                                modifier = Modifier.weight(1f),
                                labelColor = Color.White.copy(alpha = 0.85f),
                                showLabels = false,
                                showCornerIcons = true
                            )
                        }
                    }
                }
            }

            HorizontalDivider(
                modifier = Modifier.padding(horizontal = 4.dp, vertical = 10.dp),
                color = GoldPrimary.copy(alpha = 0.32f),
                thickness = 0.7.dp
            )
            Text(
                text = "พร้อมใช้ข้อมูลและสูตรนี้เพื่อจัดอันดับ",
                modifier = Modifier.fillMaxWidth(),
                textAlign = androidx.compose.ui.text.style.TextAlign.Center,
                color = GoldLight.copy(alpha = 0.76f),
                fontSize = 11.sp,
                fontWeight = FontWeight.Medium
            )
            Spacer(modifier = Modifier.height(6.dp))

            // Action Search Button (Magical Design)
            val infiniteTransition = rememberInfiniteTransition(label = "pulse")
            val pulseScale by infiniteTransition.animateFloat(
                initialValue = 1f,
                targetValue = 1.02f,
                animationSpec = infiniteRepeatable(
                    animation = tween(1200, easing = FastOutSlowInEasing),
                    repeatMode = RepeatMode.Reverse
                ),
                label = "scale"
            )
            val glowAlpha by infiniteTransition.animateFloat(
                initialValue = 0.3f,
                targetValue = 0.8f,
                animationSpec = infiniteRepeatable(
                    animation = tween(1200, easing = FastOutSlowInEasing),
                    repeatMode = RepeatMode.Reverse
                ),
                label = "glow"
            )
            val shimmerShift by infiniteTransition.animateFloat(
                initialValue = -500f,
                targetValue = 1200f,
                animationSpec = infiniteRepeatable(
                    animation = tween(2200, easing = LinearOutSlowInEasing),
                    repeatMode = RepeatMode.Restart
                ),
                label = "golden_shimmer"
            )

            // Validate the exact meaning state that searchNames sends as wish_query.
            val hasMeaningSelection = hasUserSelectedWishMeaning && pinnedWishMeaning.isNotBlank()

            val isSearching = uiState is com.miracle.names.viewmodel.UiState.Loading
            val isFormValid = !isSearching && step1Done && dayOfBirth.isNotEmpty() && hasMeaningSelection

            val buttonText = when {
                isSearching -> "กำลังจัดอันดับจากข้อมูลทั้งหมด…"
                !step1Done -> when {
                    searchMode == "standalone" -> "⚠️ โปรดพิมพ์ชื่อ"
                    isFindSurnameMode -> "⚠️ โปรดพิมพ์ชื่อตัว"
                    else -> "⚠️ โปรดพิมพ์นามสกุล"
                }
                dayOfBirth.isEmpty() -> "⚠️ โปรดเลือกวันเกิด"
                !hasMeaningSelection -> "⚠️ โปรดเลือกความหมาย"
                else -> "ดูชื่อที่ดีที่สุดของคุณ"
            }

            Button(
                onClick = { viewModel.searchNames() },
                enabled = isFormValid,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(64.dp)
                    .graphicsLayer {
                        if (isFormValid) {
                            scaleX = pulseScale
                            scaleY = pulseScale
                            shadowElevation = 18f * glowAlpha
                            shape = RoundedCornerShape(18.dp)
                            clip = false
                        }
                    },
                shape = RoundedCornerShape(18.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = Color.Transparent,
                    disabledContainerColor = Color.LightGray
                ),
                contentPadding = PaddingValues(0.dp)
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(
                            brush = Brush.linearGradient(
                                colors = if (isFormValid) listOf(
                                    Color(0xFFB88A1F),
                                    Color(0xFFD7B23A),
                                    Color(0xFFE6C85A),
                                    Color(0xFFD0A12B),
                                    Color(0xFFB88A1F)
                                ) else listOf(Color.Transparent, Color.Transparent),
                            start = Offset(shimmerShift, 0f),
                            end = Offset(shimmerShift + 800f, 0f)
                        ),
                            shape = RoundedCornerShape(18.dp)
                        )
                        .border(
                            width = 1.dp,
                            color = if (isFormValid) GoldLight.copy(alpha = 0.9f) else Color.Transparent,
                            shape = RoundedCornerShape(18.dp)
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        if (isSearching) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(18.dp),
                                strokeWidth = 2.dp,
                                color = Color.DarkGray
                            )
                            Spacer(modifier = Modifier.width(10.dp))
                        } else if (isFormValid) {
                            Icon(
                                imageVector = Icons.Default.AutoAwesome,
                                contentDescription = null,
                                tint = Color.Black.copy(alpha = 0.78f),
                                modifier = Modifier.size(18.dp)
                            )
                            Spacer(modifier = Modifier.width(10.dp))
                        }
                        Text(
                            text = buttonText,
                            color = if (isFormValid) Color.Black else Color.DarkGray,
                            fontWeight = FontWeight.ExtraBold,
                            fontSize = if (isFormValid) 17.sp else 15.sp,
                            letterSpacing = 0.5.sp
                        )
                        if (isFormValid) {
                            Spacer(modifier = Modifier.width(10.dp))
                            Icon(
                                imageVector = Icons.Default.AutoAwesome,
                                contentDescription = null,
                                tint = Color.Black.copy(alpha = 0.72f),
                                modifier = Modifier
                                    .size(16.dp)
                                    .graphicsLayer {
                                        scaleX = pulseScale
                                        scaleY = pulseScale
                                    }
                            )
                        }
                    }
                }
            }
            }
            }
            }
            }
            Spacer(modifier = Modifier.height(28.dp))
        }

        val selectedPairDetail by viewModel.selectedPairDetail.collectAsState()
        val isPairDetailLoading by viewModel.isPairDetailLoading.collectAsState()

        com.miracle.names.ui.components.PairDetailSheet(
            pairDetail = selectedPairDetail,
            isLoading = isPairDetailLoading,
            onDismiss = viewModel::dismissPairDetail
        )

        if (showResetConfirmDialog) {
            AlertDialog(
                onDismissRequest = { showResetConfirmDialog = false },
                title = {
                    Text(
                        text = "รีเซ็ตข้อมูลทั้งหมด?",
                        fontWeight = FontWeight.Bold,
                        fontSize = 17.sp,
                        color = GoldLight
                    )
                },
                text = {
                    Text(
                        text = "ข้อมูลชื่อ วันเกิด ความหมาย และผลลัพธ์จะถูกล้างเพื่อเริ่มต้นใหม่",
                        fontSize = 14.sp,
                        color = Color.White.copy(alpha = 0.85f)
                    )
                },
                confirmButton = {
                    TextButton(
                        onClick = {
                            showResetConfirmDialog = false
                            viewModel.resetAll()
                            coroutineScope.launch {
                                scrollState.animateScrollTo(0)
                            }
                        }
                    ) {
                        Text(
                            text = "รีเซ็ต",
                            color = Color(0xFFFF5252),
                            fontWeight = FontWeight.Bold
                        )
                    }
                },
                dismissButton = {
                    TextButton(
                        onClick = { showResetConfirmDialog = false }
                    ) {
                        Text(
                            text = "ยกเลิก",
                            color = Color.White.copy(alpha = 0.7f)
                        )
                    }
                },
                containerColor = Color(0xFF26160E),
                tonalElevation = 6.dp
            )
        }
    }

}
}
}


@Composable
fun SelectableChip(
    selected: Boolean,
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Box(
        modifier = modifier
            .clip(RoundedCornerShape(20.dp))
            .background(if (selected) GoldPrimary else MaterialTheme.colorScheme.surface)
            .border(
                width = 1.dp,
                color = if (selected) GoldPrimary else Color.LightGray,
                shape = RoundedCornerShape(20.dp)
            )
            .clickable { onClick() }
            .padding(horizontal = 12.dp, vertical = 8.dp),
        contentAlignment = Alignment.Center
    ) {
        Text(
            text = text,
            color = if (selected) Color.Black else MaterialTheme.colorScheme.onSurface,
            fontSize = 13.sp,
            fontWeight = if (selected) FontWeight.Bold else FontWeight.Normal
        )
    }
}

@Composable
private fun WizardWeightPercentStack(
    numerologyPercent: Int,
    semanticPercent: Int,
    soundPercent: Int,
    similarityPercent: Int,
    modifier: Modifier = Modifier
) {
    Column(
        modifier = modifier,
        verticalArrangement = Arrangement.spacedBy(6.dp)
    ) {
        WizardWeightPercentLine(
            icon = "⭐",
            label = "เลขศาสตร์",
            percent = numerologyPercent,
            accentColor = Color(0xFFFDE047)
        )
        WizardWeightPercentLine(
            icon = "🎯",
            label = "ความหมาย",
            percent = semanticPercent,
            accentColor = Color(0xFFF2B94B)
        )
        WizardWeightPercentLine(
            icon = "🔍",
            label = "เอกลักษณ์",
            percent = similarityPercent,
            accentColor = Color(0xFF4DB6AC)
        )
        WizardWeightPercentLine(
            icon = "🎶",
            label = "ไพเราะ",
            percent = soundPercent,
            accentColor = Color(0xFF7DD3A8)
        )
    }
}

@Composable
private fun WizardWeightPercentLine(
    icon: String,
    label: String,
    percent: Int,
    accentColor: Color
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(3.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text = icon,
            fontSize = 13.sp,
            maxLines = 1,
            modifier = Modifier.width(18.dp)
        )
        Text(
            text = label,
            color = Color.White.copy(alpha = 0.95f),
            fontSize = 13.sp,
            fontWeight = FontWeight.Bold,
            maxLines = 1,
            softWrap = false,
            modifier = Modifier.weight(1f)
        )
        Text(
            text = "$percent%",
            color = accentColor,
            fontSize = 15.sp,
            fontWeight = FontWeight.ExtraBold,
            maxLines = 1,
            modifier = Modifier.width(34.dp)
        )
    }
}

@Composable
fun InstantDecoderUI(satSum: Int, shaSum: Int) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = Color(0xFFFDF6E3)), // Light yellowish background from image
        shape = RoundedCornerShape(8.dp),
        border = BorderStroke(2.dp, Color(0xFFFDE047))
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(vertical = 16.dp, horizontal = 24.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            // SatSum (เลขศาสตร์)
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                Box(
                    modifier = Modifier
                        .size(64.dp)
                        .background(
                            brush = Brush.linearGradient(
                                colors = listOf(Color(0xFF4ADE80), Color(0xFF16A34A))
                            ),
                            shape = CircleShape
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "$satSum",
                        color = Color(0xFF111827),
                        fontSize = 24.sp,
                        fontWeight = FontWeight.Black
                    )
                }
                Spacer(modifier = Modifier.height(12.dp))
                Text(
                    text = "เลขศาสตร์:",
                    color = Color(0xFF3D2600),
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold
                )
            }

            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                Box(
                    modifier = Modifier
                        .size(64.dp)
                        .background(
                            brush = Brush.linearGradient(
                                colors = listOf(Color(0xFF4ADE80), Color(0xFF16A34A))
                            ),
                            shape = CircleShape
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "$shaSum",
                        color = Color(0xFF111827),
                        fontSize = 24.sp,
                        fontWeight = FontWeight.Black
                    )
                }
                Spacer(modifier = Modifier.height(12.dp))
                Text(
                    text = "พลังเงา:",
                    color = Color(0xFF3D2600),
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold
                )
            }
        }
    }
}

@Composable
fun NameProfile4DCard(
    profile: com.miracle.names.data.model.NameProfile4D,
    onWeightsChanged: ((Float, Float, Float, Float) -> Unit)? = null,
    modifier: Modifier = Modifier
) {
    Column(
        modifier = modifier
            .fillMaxWidth()
            .padding(top = 4.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        HorizontalDivider(color = Color.White.copy(alpha = 0.22f), thickness = 0.6.dp)

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(text = "📊", fontSize = 13.sp)
                Spacer(modifier = Modifier.width(6.dp))
                Text(
                    text = "โปรไฟล์ชื่อดี 4D",
                    fontSize = 12.5.sp,
                    fontWeight = FontWeight.SemiBold,
                    color = Color.White
                )
            }
            Text(
                text = if (onWeightsChanged != null) "แตะ/ลากเพื่อเลือกความสำคัญ ✨" else "ค่าประเมินเบื้องต้น",
                fontSize = 11.sp,
                fontWeight = FontWeight.Normal,
                color = if (onWeightsChanged != null) GoldLight else Color.White.copy(alpha = 0.80f)
            )
        }

        val numRaw = profile.numerologyScore.coerceIn(0f, 100f)
        val semRaw = profile.meaningScore.coerceIn(0f, 100f)
        val resRaw = profile.resonanceScore.coerceIn(0f, 100f)
        val sndRaw = profile.soundScore.coerceIn(0f, 100f)

        val total = (numRaw + semRaw + resRaw + sndRaw).coerceAtLeast(0.001f)
        val baseWeights = listOf(numRaw / total, semRaw / total, sndRaw / total, resRaw / total)

        var currentWeights by remember(numRaw, semRaw, resRaw, sndRaw) {
            mutableStateOf(baseWeights)
        }

        val numPct = (currentWeights[0] * 100f).roundToInt()
        val semPct = (currentWeights[1] * 100f).roundToInt()
        val sndPct = (currentWeights[2] * 100f).roundToInt()
        val resPct = (100 - numPct - semPct - sndPct).coerceAtLeast(0)

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(vertical = 2.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Column(
                verticalArrangement = Arrangement.spacedBy(4.dp)
            ) {
                NameProfileDimensionRow(
                    icon = "⭐",
                    name = "เลขศาสตร์",
                    score = numPct,
                    accentColor = Color(0xFFFDE047)
                )
                NameProfileDimensionRow(
                    icon = "🎯",
                    name = "ความหมาย",
                    score = semPct,
                    accentColor = Color(0xFFF2B94B)
                )
                NameProfileDimensionRow(
                    icon = "🔍",
                    name = "เอกลักษณ์",
                    score = resPct,
                    accentColor = Color(0xFF4DB6AC)
                )
                NameProfileDimensionRow(
                    icon = "🎶",
                    name = "ไพเราะ",
                    score = sndPct,
                    accentColor = Color(0xFF7DD3A8)
                )
            }

            NameProfileRadarView(
                wNum = currentWeights[0],
                wSem = currentWeights[1],
                wRes = currentWeights[3],
                wSnd = currentWeights[2],
                onWeightsChanged = { n, sem, snd, sim ->
                    currentWeights = listOf(n, sem, snd, sim)
                    onWeightsChanged?.invoke(n, sem, snd, sim)
                },
                modifier = Modifier
                    .weight(1f)
                    .height(145.dp)
                    .padding(start = 8.dp)
            )
        }
    }
}

@Composable
private fun NameProfileDimensionRow(
    icon: String,
    name: String,
    score: Int,
    accentColor: Color
) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(6.dp)
    ) {
        Text(
            text = icon,
            fontSize = 13.sp
        )

        Text(
            text = name,
            fontSize = 12.sp,
            fontWeight = FontWeight.Medium,
            color = Color.White,
            modifier = Modifier.width(58.dp)
        )

        Text(
            text = "$score%",
            fontSize = 12.5.sp,
            fontWeight = FontWeight.Bold,
            color = accentColor,
            modifier = Modifier.width(36.dp)
        )
    }
}

@Composable
private fun NameProfileRadarView(
    wNum: Float,
    wSem: Float,
    wRes: Float,
    wSnd: Float,
    onWeightsChanged: ((Float, Float, Float, Float) -> Unit)? = null,
    modifier: Modifier = Modifier
) {
    var activeWeights by remember(wNum, wSem, wRes, wSnd) {
        mutableStateOf(listOf(wNum, wSem, wSnd, wRes))
    }

    val currentOnWeightsChanged by rememberUpdatedState(onWeightsChanged)

    val dragModifier = if (onWeightsChanged != null) {
        modifier.pointerInput(Unit) {
            fun update(position: androidx.compose.ui.geometry.Offset) {
                val weights = com.miracle.names.ui.components.diamondWeightsAt(
                    position = position,
                    width = size.width.toFloat(),
                    height = size.height.toFloat(),
                    horizontalInset = 18.dp.toPx(),
                    verticalInset = 18.dp.toPx()
                )
                activeWeights = weights
                currentOnWeightsChanged?.invoke(weights[0], weights[1], weights[2], weights[3])
            }
            awaitEachGesture {
                val down = awaitFirstDown(requireUnconsumed = false)
                update(down.position)
                down.consume()
                while (true) {
                    val event = awaitPointerEvent()
                    val change = event.changes.firstOrNull { it.id == down.id } ?: break
                    if (!change.pressed) break
                    update(change.position)
                    change.consume()
                }
            }
        }
    } else {
        modifier
    }

    Box(modifier = dragModifier) {
        androidx.compose.foundation.Canvas(modifier = Modifier.fillMaxSize()) {
            val w = size.width
            val h = size.height
            val cx = w / 2f
            val cy = h / 2f
            val rx = (w / 2f) - 18.dp.toPx()
            val ry = (h / 2f) - 18.dp.toPx()

            val top = androidx.compose.ui.geometry.Offset(cx, cy - ry)
            val right = androidx.compose.ui.geometry.Offset(cx + rx, cy)
            val bottom = androidx.compose.ui.geometry.Offset(cx, cy + ry)
            val left = androidx.compose.ui.geometry.Offset(cx - rx, cy)

            val diamondPath = androidx.compose.ui.graphics.Path().apply {
                moveTo(top.x, top.y)
                lineTo(right.x, right.y)
                lineTo(bottom.x, bottom.y)
                lineTo(left.x, left.y)
                close()
            }

            // 1. Diamond Gradient Fill
            drawPath(
                path = diamondPath,
                brush = androidx.compose.ui.graphics.Brush.verticalGradient(
                    listOf(
                        Color(0xFFFDE047).copy(alpha = 0.32f),
                        Color(0xFF382014).copy(alpha = 0.50f)
                    )
                )
            )

            // 2. Diamond Frame Stroke
            drawPath(
                path = diamondPath,
                brush = androidx.compose.ui.graphics.Brush.linearGradient(
                    listOf(
                        Color(0xFFFDE047),
                        Color(0xFF84CC16)
                    )
                ),
                style = androidx.compose.ui.graphics.drawscope.Stroke(width = 2.dp.toPx())
            )

            // 3. Crosshair Lines
            drawLine(
                color = Color.White.copy(alpha = 0.22f),
                start = top,
                end = bottom,
                strokeWidth = 1.2.dp.toPx()
            )
            drawLine(
                color = Color.White.copy(alpha = 0.22f),
                start = left,
                end = right,
                strokeWidth = 1.2.dp.toPx()
            )

            // 4. Corner Dots
            val cornerR = 3.dp.toPx()
            drawCircle(color = Color(0xFFFDE047), radius = cornerR, center = top)
            drawCircle(color = Color(0xFF4DB6AC), radius = cornerR, center = right)
            drawCircle(color = Color(0xFF7DD3A8), radius = cornerR, center = bottom)
            drawCircle(color = Color(0xFFF2B94B), radius = cornerR, center = left)

            // 5. Spark Indicator & Halo
            val currentNum = activeWeights.getOrElse(0) { wNum }
            val currentSem = activeWeights.getOrElse(1) { wSem }
            val currentSnd = activeWeights.getOrElse(2) { wSnd }
            val currentRes = activeWeights.getOrElse(3) { wRes }

            val sparkCenter = androidx.compose.ui.geometry.Offset(
                x = (top.x * currentNum) + (bottom.x * currentSnd) + (left.x * currentSem) + (right.x * currentRes),
                y = (top.y * currentNum) + (bottom.y * currentSnd) + (left.y * currentSem) + (right.y * currentRes)
            )

            drawCircle(
                brush = androidx.compose.ui.graphics.Brush.radialGradient(
                    colors = listOf(
                        Color(0xFFFFF8D5).copy(alpha = 0.90f),
                        Color(0xFFFFC84A).copy(alpha = 0.45f),
                        Color.Transparent
                    ),
                    center = sparkCenter,
                    radius = 18.dp.toPx()
                ),
                radius = 18.dp.toPx(),
                center = sparkCenter
            )
            drawCircle(
                color = Color(0xFFFFD65A).copy(alpha = 0.55f),
                radius = 11.dp.toPx(),
                center = sparkCenter,
                style = androidx.compose.ui.graphics.drawscope.Stroke(width = 1.5.dp.toPx())
            )
            drawCircle(
                color = Color.White,
                radius = 6.dp.toPx(),
                center = sparkCenter
            )
            drawCircle(
                color = Color(0xFFF59E0B),
                radius = 3.2.dp.toPx(),
                center = sparkCenter
            )
        }

        // Corner Icons
        Text(
            text = "⭐",
            fontSize = 13.sp,
            modifier = Modifier
                .align(Alignment.TopCenter)
                .offset(y = 1.dp)
        )
        Text(
            text = "🎯",
            fontSize = 13.sp,
            modifier = Modifier
                .align(Alignment.CenterStart)
                .offset(x = 1.dp)
        )
        Text(
            text = "🔍",
            fontSize = 13.sp,
            modifier = Modifier
                .align(Alignment.CenterEnd)
                .offset(x = (-1).dp)
        )
        Text(
            text = "🎶",
            fontSize = 13.sp,
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .offset(y = (-1).dp)
        )
    }
}

@Composable
fun WizardStepBar(
    currentStep: Int,
    step1Done: Boolean,
    step2Done: Boolean,
    step3Done: Boolean,
    step4Done: Boolean,
    onReset: (() -> Unit)? = null,
    onStepClick: ((Int) -> Unit)? = null
) {
    Surface(
        shape = RoundedCornerShape(14.dp),
        color = Color(0xFF1F120B),
        border = BorderStroke(1.dp, GoldPrimary.copy(alpha = 0.22f)),
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 4.dp, vertical = 2.dp)
    ) {
        Column(
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 10.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                    Icon(
                        imageVector = Icons.Default.AutoAwesome,
                        contentDescription = null,
                        tint = GoldPrimary,
                        modifier = Modifier.size(14.dp)
                    )
                    Text(
                        text = "4 ขั้นตอนการตั้งชื่อมงคล",
                        fontSize = 12.5.sp,
                        fontWeight = FontWeight.Bold,
                        color = GoldLight
                    )
                }
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(7.dp)
                ) {
                    Text(
                        text = "ขั้นตอนที่ $currentStep / 4",
                        fontSize = 11.sp,
                        fontWeight = FontWeight.SemiBold,
                        color = GoldPrimary
                    )

                    if (onReset != null) {
                        Surface(
                            onClick = onReset,
                            shape = RoundedCornerShape(50),
                            color = Color.White.copy(alpha = 0.08f),
                            border = BorderStroke(0.8.dp, GoldPrimary.copy(alpha = 0.32f))
                        ) {
                            Row(
                                modifier = Modifier.padding(horizontal = 7.dp, vertical = 2.5.dp),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(3.5.dp)
                            ) {
                                Icon(
                                    imageVector = Icons.Default.Refresh,
                                    contentDescription = null,
                                    tint = GoldPrimary,
                                    modifier = Modifier.size(11.dp)
                                )
                                Text(
                                    text = "รีเซ็ต",
                                    fontSize = 10.5.sp,
                                    fontWeight = FontWeight.Bold,
                                    color = GoldPrimary
                                )
                            }
                        }
                    }
                }
            }

            // Pulse animation for the active step beacon
            val pulseTransition = rememberInfiniteTransition(label = "step_beacon_pulse")
            val beaconPulse by pulseTransition.animateFloat(
                initialValue = 0f,
                targetValue = 1f,
                animationSpec = infiniteRepeatable(
                    animation = tween(1100, easing = FastOutSlowInEasing),
                    repeatMode = RepeatMode.Reverse
                ),
                label = "step_beacon_pulse_val"
            )
            val lineFlow by pulseTransition.animateFloat(
                initialValue = -1f,
                targetValue = 2f,
                animationSpec = infiniteRepeatable(
                    animation = tween(1250, easing = LinearEasing),
                    repeatMode = RepeatMode.Restart
                ),
                label = "step_connector_flow"
            )

            // One visual language per state: completed, current, and upcoming.
            // Keeping the glow exclusive to the current step makes the next action
            // scannable without turning the whole progress bar into a light source.
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .drawWithContent {
                        val y = 16.dp.toPx()
                        val radius = 10.dp.toPx()
                        val centers = List(4) { size.width * (it + 0.5f) / 4f }
                        val completedSegments = listOf(step1Done, step2Done, step3Done).takeWhile { it }.size
                        repeat(3) { index ->
                            val start = Offset(centers[index] + radius, y)
                            val end = Offset(centers[index + 1] - radius, y)
                            drawLine(Color(0xFF1F120B), start, end, strokeWidth = 5.dp.toPx())
                            drawLine(Color.White.copy(alpha = 0.12f), start, end, strokeWidth = 2.5.dp.toPx())
                            if (index < completedSegments) {
                                drawLine(GoldPrimary, start, end, strokeWidth = 2.5.dp.toPx())
                            }
                        }
                        if (completedSegments > 0) {
                            // Match iOS: animate only the next action path. For example,
                            // after Steps 1 and 2 are done, travel from Step 2 to Step 3.
                            val trackStart = centers[completedSegments - 1] + radius
                            val trackEnd = centers[completedSegments] - radius
                            val travel = ((lineFlow + 1f) / 3f).coerceIn(0f, 1f)
                            val cometX = trackStart + ((trackEnd - trackStart) * travel)
                            drawLine(
                                color = GoldLight.copy(alpha = 0.30f),
                                start = Offset((cometX - 18.dp.toPx()).coerceAtLeast(trackStart), y),
                                end = Offset(cometX, y),
                                strokeWidth = 3.dp.toPx()
                            )
                            drawCircle(GoldLight, radius = 3.dp.toPx(), center = Offset(cometX, y))
                        }
                        drawContent()
                    },
                verticalAlignment = Alignment.Top,
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                val steps = listOf(
                    Triple(1, "ระบุชื่อ", step1Done),
                    Triple(2, "วันเกิด", step2Done),
                    Triple(3, "ความหมาย", step3Done),
                    Triple(4, "ค้นหา 4 มิติ", step4Done)
                )

                val completedGoldBrush = Brush.linearGradient(
                    listOf(Color(0xFFFFF0A3), Color(0xFFD4AF37), Color(0xFFA37C10))
                )
                val completedLineBrush = Brush.horizontalGradient(
                    listOf(GoldPrimary, Color(0xFFFFF0A3), GoldPrimary)
                )
                val lineFlowBrush = Brush.linearGradient(
                    colors = listOf(
                        Color.Transparent,
                        Color(0xFFFFF0A3).copy(alpha = 0.70f),
                        Color.Transparent
                    ),
                    start = Offset(lineFlow * 160f, 0f),
                    end = Offset((lineFlow + 1f) * 160f, 0f)
                )
                val activeGold = GoldPrimary
                val inactiveLineColor = Color.White.copy(alpha = 0.12f)

                steps.forEachIndexed { index, (stepNum, title, isDone) ->
                    val isActive = currentStep == stepNum
                    val prevDone = if (index > 0) steps[index - 1].third else false

                    val leftLineProgress by animateFloatAsState(
                        targetValue = if (prevDone) 1f else 0f,
                        animationSpec = tween(durationMillis = 420, easing = FastOutSlowInEasing),
                        label = "step_left_line_$stepNum"
                    )
                    val rightLineProgress by animateFloatAsState(
                        targetValue = if (isDone) 1f else 0f,
                        animationSpec = tween(durationMillis = 420, easing = FastOutSlowInEasing),
                        label = "step_right_line_$stepNum"
                    )

                    val nodeScale by animateFloatAsState(
                        targetValue = when {
                            isDone -> 1.0f
                            isActive -> 1.05f + (0.05f * beaconPulse)
                            else -> 0.94f
                        },
                        animationSpec = spring(
                            dampingRatio = Spring.DampingRatioMediumBouncy,
                            stiffness = Spring.StiffnessMediumLow
                        ),
                        label = "step_scale_$stepNum"
                    )

                    val titleColor by animateColorAsState(
                        targetValue = when {
                            isDone -> Color(0xFFFFDF85)
                            isActive -> activeGold
                            else -> Color.White.copy(alpha = 0.45f)
                        },
                        animationSpec = tween(durationMillis = 300),
                        label = "step_title_color_$stepNum"
                    )

                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        modifier = Modifier
                            .weight(1f)
                            .clickable(enabled = onStepClick != null) {
                                onStepClick?.invoke(stepNum)
                            }
                            .padding(vertical = 2.dp)
                    ) {
                        Box(
                            // Every step reserves the halo's full height. This keeps
                            // connector segments on one shared horizontal centerline.
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(34.dp),
                            contentAlignment = Alignment.Center
                        ) {
                            if (isActive) {
                                val haloAlpha = 0.50f + (0.45f * beaconPulse)
                                // Layer 1: Wide radiant golden bloom (42.dp)
                                Box(
                                    modifier = Modifier
                                        .requiredSize(42.dp)
                                        .background(
                                            Brush.radialGradient(
                                                colorStops = arrayOf(
                                                    0.0f to Color(0xFFFFDF85).copy(alpha = haloAlpha),
                                                    0.40f to Color(0xFFE5B842).copy(alpha = haloAlpha * 0.90f),
                                                    0.65f to GoldPrimary.copy(alpha = haloAlpha * 0.65f),
                                                    0.85f to GoldPrimary.copy(alpha = haloAlpha * 0.25f),
                                                    1.0f to Color.Transparent
                                                )
                                            ),
                                            CircleShape
                                        )
                                )
                                // Layer 2: Glistening corona edge ring (28.dp, directly hugging the 24.dp node)
                                Box(
                                    modifier = Modifier
                                        .size(28.dp)
                                        .border(
                                            width = 1.6.dp,
                                            color = Color(0xFFFFF0A3).copy(alpha = 0.85f * beaconPulse),
                                            shape = CircleShape
                                        )
                                )
                            } else if (isDone) {
                                Box(
                                    modifier = Modifier
                                        .requiredSize(38.dp)
                                        .background(
                                            Brush.radialGradient(
                                                colorStops = arrayOf(
                                                    0.0f to GoldPrimary.copy(alpha = 0.55f),
                                                    0.45f to Color(0xFFFFE082).copy(alpha = 0.35f),
                                                    0.75f to GoldPrimary.copy(alpha = 0.18f),
                                                    1.0f to Color.Transparent
                                                )
                                            ),
                                            CircleShape
                                        )
                                )
                            }

                            // Central Node (Fixed 24dp, rock-solid geometry)
                            Box(
                                modifier = Modifier
                                    .size(24.dp)
                                    .then(
                                        when {
                                            isDone -> Modifier.background(completedGoldBrush, CircleShape)
                                            isActive -> Modifier.background(
                                                Brush.radialGradient(
                                                    listOf(Color(0xFF5A351D), Color(0xFF2B1A0E))
                                                ),
                                                CircleShape
                                            )
                                            else -> Modifier.background(Color(0xFF26160E), CircleShape)
                                        }
                                    )
                                    .border(
                                        width = if (isDone) 1.5.dp else if (isActive) 1.8.dp else 1.dp,
                                        color = when {
                                            isDone -> Color(0xFFFFF7D6)
                                            isActive -> Color(0xFFFFF0A3)
                                            else -> Color.White.copy(alpha = 0.18f)
                                        },
                                        shape = CircleShape
                                    ),
                                contentAlignment = Alignment.Center
                            ) {
                                AnimatedContent(
                                    targetState = isDone,
                                    transitionSpec = {
                                        (fadeIn(tween(200)) + scaleIn(spring(dampingRatio = Spring.DampingRatioMediumBouncy, stiffness = Spring.StiffnessMediumLow)))
                                            .togetherWith(fadeOut(tween(150)) + scaleOut())
                                    },
                                    label = "step_node_content_$stepNum"
                                ) { done ->
                                    if (done) {
                                        Icon(
                                            imageVector = Icons.Default.Check,
                                            contentDescription = null,
                                            tint = Color(0xFF26160E),
                                            modifier = Modifier.size(14.dp)
                                        )
                                    } else {
                                        Text(
                                            text = "$stepNum",
                                            fontSize = if (isActive) 11.5.sp else 11.sp,
                                            fontWeight = if (isActive) FontWeight.ExtraBold else FontWeight.Bold,
                                            color = if (isActive) Color(0xFFFFF7D6) else Color.White.copy(alpha = 0.52f)
                                        )
                                    }
                                }
                            }
                        }
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(
                            text = title,
                            fontSize = 10.sp,
                            fontWeight = if (isActive || isDone) FontWeight.Bold else FontWeight.Normal,
                            color = titleColor,
                            maxLines = 1,
                            textAlign = androidx.compose.ui.text.style.TextAlign.Center
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun MagicSuggestionsSkeleton(
    magicGlowAlpha: Float,
    magicRotation: Float,
    modifier: Modifier = Modifier
) {
    val transition = rememberInfiniteTransition(label = "magic_shimmer")
    val translateAnim by transition.animateFloat(
        initialValue = -500f,
        targetValue = 1200f,
        animationSpec = infiniteRepeatable(
            animation = tween(durationMillis = 1350, easing = androidx.compose.animation.core.LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "shimmer_translate"
    )
    val shimmerBrush = Brush.linearGradient(
        colors = listOf(
            Color(0xFF1F120B),
            Color(0xFF382014),
            GoldPrimary.copy(alpha = 0.42f),
            Color(0xFF382014),
            Color(0xFF1F120B)
        ),
        start = Offset(translateAnim, translateAnim * 0.35f),
        end = Offset(translateAnim + 360f, translateAnim * 0.35f + 220f)
    )

    Column(
        modifier = modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(6.dp),
            modifier = Modifier.padding(horizontal = 2.dp, vertical = 2.dp)
        ) {
            Icon(
                imageVector = Icons.Default.AutoAwesome,
                contentDescription = null,
                tint = GoldPrimary,
                modifier = Modifier
                    .size(13.dp)
                    .graphicsLayer {
                        val scale = 0.85f + (0.3f * magicGlowAlpha)
                        scaleX = scale
                        scaleY = scale
                        rotationZ = magicRotation
                    }
            )
            Text(
                text = "กำลังคำนวณและตรวจคู่เลขชื่อใกล้เคียง...",
                fontSize = 12.sp,
                fontWeight = FontWeight.SemiBold,
                color = GoldLight.copy(alpha = magicGlowAlpha)
            )
        }

        val rowWidths = listOf(86.dp, 64.dp, 98.dp)
        rowWidths.forEach { nameWidth ->
            Surface(
                shape = RoundedCornerShape(10.dp),
                color = Color(0xFF26160E),
                border = BorderStroke(0.7.dp, GoldPrimary.copy(alpha = 0.22f)),
                modifier = Modifier.fillMaxWidth()
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 12.dp, vertical = 8.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Box(
                        modifier = Modifier
                            .size(14.dp)
                            .clip(RoundedCornerShape(3.dp))
                            .background(shimmerBrush)
                    )
                    Spacer(modifier = Modifier.width(10.dp))
                    Box(
                        modifier = Modifier
                            .width(nameWidth)
                            .height(13.dp)
                            .clip(RoundedCornerShape(4.dp))
                            .background(shimmerBrush)
                    )

                    Spacer(modifier = Modifier.weight(1f))

                    Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                        Box(
                            modifier = Modifier
                                .size(28.dp)
                                .clip(CircleShape)
                                .background(shimmerBrush)
                                .border(1.dp, GoldPrimary.copy(alpha = 0.35f), CircleShape)
                        )
                        Box(
                            modifier = Modifier
                                .size(28.dp)
                                .clip(CircleShape)
                                .background(shimmerBrush)
                                .border(1.dp, GoldPrimary.copy(alpha = 0.35f), CircleShape)
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun StepHeaderPattern(
    isCompleted: Boolean = false,
    modifier: Modifier = Modifier
) {
    androidx.compose.foundation.Canvas(modifier = modifier) {
        val step = 22.dp.toPx()
        val halfStep = step / 2f
        val totalW = size.width
        val totalH = size.height
        if (totalW <= 0f || totalH <= 0f) return@Canvas

        val gold = GoldPrimary
        val lineAlpha = if (isCompleted) 0.13f else 0.08f
        val starAlpha = if (isCompleted) 0.22f else 0.14f
        val dotAlpha  = if (isCompleted) 0.16f else 0.10f
        val strokeWidth = 1.0.dp.toPx()

        // 1. Diagonal grid lines anchored to origin (0, 0)
        val minK = kotlin.math.floor(-totalH / step).toInt() - 1
        val maxK = kotlin.math.ceil(totalW / step).toInt() + 1
        for (k in minK..maxK) {
            val x0 = k * step
            drawLine(
                color = gold.copy(alpha = lineAlpha),
                start = Offset(x0, 0f),
                end = Offset(x0 + totalH, totalH),
                strokeWidth = strokeWidth
            )
        }

        val minM = -1
        val maxM = kotlin.math.ceil((totalW + totalH) / step).toInt() + 1
        for (m in minM..maxM) {
            val x0 = m * step
            drawLine(
                color = gold.copy(alpha = lineAlpha),
                start = Offset(x0, 0f),
                end = Offset(x0 - totalH, totalH),
                strokeWidth = strokeWidth
            )
        }

        // 2. Intersections (Thai 4-pointed Pra-jam-yam star) & cell centers (dot)
        val starR = 3.2.dp.toPx()
        val innerR = 1.2.dp.toPx()
        val dotR = 1.1.dp.toPx()

        val maxRow = kotlin.math.ceil(totalH / halfStep).toInt() + 1
        val minCol = -1
        val maxCol = kotlin.math.ceil(totalW / step).toInt() + 1

        val starPath = androidx.compose.ui.graphics.Path()

        for (row in 0..maxRow) {
            val cy = row * halfStep
            val isOdd = (row % 2 != 0)

            // Intersections
            val xOffset = if (isOdd) halfStep else 0f
            for (col in minCol..maxCol) {
                val cx = col * step + xOffset
                if (cx >= -starR && cx <= totalW + starR) {
                    starPath.reset()
                    starPath.moveTo(cx, cy - starR)
                    starPath.lineTo(cx + innerR, cy - innerR)
                    starPath.lineTo(cx + starR, cy)
                    starPath.lineTo(cx + innerR, cy + innerR)
                    starPath.lineTo(cx, cy + starR)
                    starPath.lineTo(cx - innerR, cy + innerR)
                    starPath.lineTo(cx - starR, cy)
                    starPath.lineTo(cx - innerR, cy - innerR)
                    starPath.close()
                    drawPath(starPath, color = gold.copy(alpha = starAlpha))
                }
            }

            // Cell centers
            val dotOffset = if (isOdd) 0f else halfStep
            for (col in minCol..maxCol) {
                val dotX = col * step + dotOffset
                if (dotX >= -dotR && dotX <= totalW + dotR) {
                    drawCircle(
                        color = gold.copy(alpha = dotAlpha),
                        radius = dotR,
                        center = Offset(dotX, cy)
                    )
                }
            }
        }
    }
}

@Composable
fun WizardStepHeading(
    number: String,
    title: String,
    subtitle: String? = null,
    isCompleted: Boolean = false,
    isActive: Boolean = false,
    integratedWithSection: Boolean = false,
    modifier: Modifier = Modifier
) {
    val guidanceTransition = rememberInfiniteTransition(label = "section_guidance_$number")
    val guidancePulse by guidanceTransition.animateFloat(
        initialValue = 0f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            animation = tween(900, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "section_guidance_pulse_$number"
    )
    val isGuidanceTarget = isActive && !isCompleted
    Surface(
        shape = RoundedCornerShape(if (integratedWithSection) 0.dp else 14.dp),
        color = if (isCompleted) Color(0xFF382014) else Color(0xFF26160E),
        border = if (integratedWithSection) null else BorderStroke(
            1.dp,
            if (isCompleted) GoldPrimary.copy(alpha = 0.55f) else GoldPrimary.copy(alpha = 0.28f)
        ),
        modifier = modifier
            .padding(horizontal = if (integratedWithSection) 8.dp else 0.dp)
            .fillMaxWidth()
    ) {
        Box(modifier = Modifier.fillMaxWidth()) {
            StepHeaderPattern(
                isCompleted = isCompleted,
                modifier = Modifier.matchParentSize()
            )
            Column {
                Row(
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Box(
                        modifier = Modifier.size(36.dp),
                        contentAlignment = Alignment.Center
                    ) {
                        if (isGuidanceTarget) {
                            val haloAlpha = 0.50f + (0.45f * guidancePulse)
                            // Layer 1: Wide radiant golden bloom (48.dp) with color stops peaking right outside the node
                            Box(
                                modifier = Modifier
                                    .requiredSize(48.dp)
                                    .background(
                                        Brush.radialGradient(
                                            colorStops = arrayOf(
                                                0.0f to Color(0xFFFFDF85).copy(alpha = haloAlpha),
                                                0.40f to Color(0xFFE5B842).copy(alpha = haloAlpha * 0.90f),
                                                0.65f to GoldPrimary.copy(alpha = haloAlpha * 0.65f),
                                                0.85f to GoldPrimary.copy(alpha = haloAlpha * 0.25f),
                                                1.0f to Color.Transparent
                                            )
                                        ),
                                        CircleShape
                                    )
                            )
                            // Layer 2: Glistening corona edge ring (32.dp, directly hugging the 28.dp node)
                            Box(
                                modifier = Modifier
                                    .size(32.dp)
                                    .border(
                                        width = 1.6.dp,
                                        color = Color(0xFFFFF0A3).copy(alpha = 0.85f * guidancePulse),
                                        shape = CircleShape
                                    )
                            )
                        } else if (isCompleted) {
                            Box(
                                modifier = Modifier
                                    .requiredSize(42.dp)
                                    .background(
                                        Brush.radialGradient(
                                            colorStops = arrayOf(
                                                0.0f to GoldPrimary.copy(alpha = 0.55f),
                                                0.45f to Color(0xFFFFE082).copy(alpha = 0.35f),
                                                0.75f to GoldPrimary.copy(alpha = 0.18f),
                                                1.0f to Color.Transparent
                                            )
                                        ),
                                        CircleShape
                                    )
                            )
                        }

                        Box(
                            modifier = Modifier
                                .size(28.dp)
                                .clip(CircleShape)
                                .then(
                                    when {
                                        isCompleted -> Modifier.background(
                                            Brush.linearGradient(
                                                listOf(Color(0xFFFFF0A3), Color(0xFFD4AF37), Color(0xFFA37C10))
                                            )
                                        )
                                        isGuidanceTarget -> Modifier.background(
                                            Brush.radialGradient(
                                                listOf(Color(0xFF5A351D), Color(0xFF2B1A0E))
                                            )
                                        )
                                        else -> Modifier.background(Color(0xFF3D2418))
                                    }
                                )
                                .border(
                                    width = if (isCompleted) 1.2.dp else if (isGuidanceTarget) 1.8.dp else 1.2.dp,
                                    color = if (isCompleted) Color(0xFFFFF7D6) else if (isGuidanceTarget) Color(0xFFFFF0A3) else GoldPrimary.copy(alpha = 0.6f),
                                    shape = CircleShape
                                ),
                            contentAlignment = Alignment.Center
                        ) {
                            if (isCompleted) {
                                Icon(
                                    imageVector = Icons.Default.Check,
                                    contentDescription = null,
                                    tint = Color(0xFF26160E),
                                    modifier = Modifier.size(16.dp)
                                )
                            } else {
                                Text(
                                    text = number,
                                    color = if (isGuidanceTarget) Color(0xFFFFF7D6) else GoldLight,
                                    fontWeight = FontWeight.Bold,
                                    fontSize = 13.sp
                                )
                            }
                        }
                    }
                    Spacer(modifier = Modifier.width(10.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Text(
                            text = title,
                            color = Color.White,
                            fontSize = 14.5.sp,
                            fontWeight = FontWeight.Bold,
                            fontFamily = PromptFontFamily
                        )
                        if (!subtitle.isNullOrBlank()) {
                            Text(
                                text = subtitle,
                                color = GoldLight.copy(alpha = 0.75f),
                                fontSize = 11.5.sp,
                                lineHeight = 15.sp
                            )
                        }
                    }
                }
                if (integratedWithSection) {
                    HorizontalDivider(
                        color = GoldPrimary.copy(alpha = 0.18f),
                        thickness = 0.75.dp
                    )
                }
            }
        }
    }
}

data class KalakiniDayDisplay(
    val dayName: String,
    val letters: String
)

fun getKalakiniDayInfo(dayKey: String): KalakiniDayDisplay? {
    return when (dayKey) {
        "sunday" -> KalakiniDayDisplay("วันอาทิตย์", "ศษสหฬฮ")
        "monday" -> KalakiniDayDisplay("วันจันทร์", "ะาิีึืุูเแโใไ")
        "tuesday" -> KalakiniDayDisplay("วันอังคาร", "กขคฆง")
        "wednesday1" -> KalakiniDayDisplay("วันพุธ (กลางวัน)", "จฉชซฌญ")
        "wednesday2" -> KalakiniDayDisplay("วันพุธ (กลางคืน)", "บปผฝพฟภม")
        "thursday" -> KalakiniDayDisplay("วันพฤหัสบดี", "ดตถทธน")
        "friday" -> KalakiniDayDisplay("วันศุกร์", "ยรลว")
        "saturday" -> KalakiniDayDisplay("วันเสาร์", "ฎฏฐฑฒณ")
        else -> null
    }
}

@Composable
fun KalakiniSummaryCard(
    dayInfo: KalakiniDayDisplay,
    modifier: Modifier = Modifier
) {
    Surface(
        shape = RoundedCornerShape(12.dp),
        color = Color(0xFF26160E),
        border = BorderStroke(1.dp, GoldPrimary.copy(alpha = 0.28f)),
        modifier = modifier
            .fillMaxWidth()
            .padding(horizontal = 10.dp)
            .padding(bottom = 9.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 12.dp, vertical = 10.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                imageVector = Icons.Default.CheckCircle,
                contentDescription = null,
                tint = Color(0xFFFFDF85),
                modifier = Modifier.size(16.dp)
            )
            Spacer(Modifier.width(8.dp))
            Text(
                text = "คัดเฉพาะชื่อมงคลปลอดกาลกิณี 100%",
                fontSize = 12.5.sp,
                fontWeight = FontWeight.Medium,
                color = GoldLight.copy(alpha = 0.92f)
            )
        }
    }
}
