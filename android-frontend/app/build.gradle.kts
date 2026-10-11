import java.util.Properties
import java.io.FileInputStream

plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.plugin.compose")
    id("com.google.gms.google-services")
}

android {
    namespace = "com.miracle.names"
    compileSdk = 36

    defaultConfig {
        applicationId = "com.numberniceic.app"
        minSdk = 26
        targetSdk = 36
        versionCode = 32
        versionName = "1.5.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
        vectorDrawables {
            useSupportLibrary = true
        }
    }

    val localProps = Properties().apply {
        val localPropsFile = rootProject.file("local.properties")
        if (localPropsFile.exists()) {
            FileInputStream(localPropsFile).use { load(it) }
        }
    }
    val releaseKeystorePath = providers.environmentVariable("CHUEDEE_KEYSTORE_PATH").orNull
        ?: localProps.getProperty("CHUEDEE_KEYSTORE_PATH")
        ?: if (file("upload-keystore.jks").exists()) "upload-keystore.jks" else null
    val releaseKeystorePassword = providers.environmentVariable("CHUEDEE_KEYSTORE_PASSWORD").orNull
        ?: localProps.getProperty("CHUEDEE_KEYSTORE_PASSWORD")
    val releaseKeyAlias = providers.environmentVariable("CHUEDEE_KEY_ALIAS").orNull
        ?: localProps.getProperty("CHUEDEE_KEY_ALIAS")
    val releaseKeyPassword = providers.environmentVariable("CHUEDEE_KEY_PASSWORD").orNull
        ?: localProps.getProperty("CHUEDEE_KEY_PASSWORD")

    signingConfigs {
        if (
            releaseKeystorePath != null &&
            releaseKeystorePassword != null &&
            releaseKeyAlias != null &&
            releaseKeyPassword != null
        ) {
            create("release") {
                storeFile = file(releaseKeystorePath)
                storePassword = releaseKeystorePassword
                keyAlias = releaseKeyAlias
                keyPassword = releaseKeyPassword
            }
        }
    }

    buildTypes {
        debug {
            buildConfigField("String", "API_BASE_URL", "\"https://xn--b3cu8e7ah6h.com/\"")
        }
        release {
            signingConfigs.findByName("release")?.let { signingConfig = it }
            buildConfigField("String", "API_BASE_URL", "\"https://xn--b3cu8e7ah6h.com/\"")
            isMinifyEnabled = false
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
    buildFeatures {
        compose = true
        buildConfig = true
    }
    packaging {
        resources {
            excludes += "/META-INF/{AL2.0,LGPL2.1}"
        }
    }
}

kotlin {
    compilerOptions {
        jvmTarget = org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17
    }
}

dependencies {
    implementation("androidx.core:core-splashscreen:1.0.1")
    implementation("io.coil-kt:coil-compose:2.6.0")
    implementation("io.coil-kt:coil-svg:2.6.0")
    implementation("com.android.billingclient:billing-ktx:8.3.0")

    // Firebase Cloud Messaging
    implementation(platform("com.google.firebase:firebase-bom:33.1.2"))
    implementation("com.google.firebase:firebase-messaging-ktx")

    implementation("androidx.core:core-ktx:1.12.0")
    implementation("androidx.lifecycle:lifecycle-runtime-ktx:2.7.0")
    implementation("androidx.activity:activity-compose:1.8.2")
    
    // Compose
    implementation(platform("androidx.compose:compose-bom:2024.02.00"))
    implementation("androidx.compose.ui:ui")
    implementation("androidx.compose.ui:ui-graphics")
    implementation("androidx.compose.ui:ui-tooling-preview")
    implementation("androidx.compose.material3:material3")
    implementation("androidx.compose.material:material-icons-extended")
    implementation("androidx.lifecycle:lifecycle-viewmodel-compose:2.7.0")
    implementation("dev.shreyaspatil:capturable:2.1.0")

    // Retrofit & Gson for Networking
    implementation("com.squareup.retrofit2:retrofit:2.9.0")
    implementation("com.squareup.retrofit2:converter-gson:2.9.0")
    implementation("com.squareup.okhttp3:logging-interceptor:4.12.0")

    // Coroutines
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-android:1.7.3")

    testImplementation("junit:junit:4.13.2")
    androidTestImplementation("androidx.test.ext:junit:1.1.5")
    androidTestImplementation("androidx.test.espresso:espresso-core:3.5.1")
    androidTestImplementation(platform("androidx.compose:compose-bom:2024.02.00"))
    androidTestImplementation("androidx.compose.ui:ui-test-junit4")
    debugImplementation("androidx.compose.ui:ui-tooling")
    debugImplementation("androidx.compose.ui:ui-test-manifest")
}
