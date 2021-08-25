import org.jetbrains.kotlin.gradle.tasks.KotlinCompile

plugins {
    kotlin("jvm") version "1.5.10"
    application
}

group = "com.ifulfilment"
version = "1.0"

repositories {
    mavenCentral()
}

tasks.withType<KotlinCompile>() {
    kotlinOptions.jvmTarget = "1.8"
}

application {
    mainClass.set("MainKt")
}

dependencies{
    implementation("com.google.code.gson:gson:2.8.2")
}

tasks.jar {
    duplicatesStrategy = DuplicatesStrategy.INCLUDE
    archiveFileName.set("printer-tools.jar")
    destinationDirectory.set(file("../libs"))
    manifest.attributes.apply {
        put("Main-Class", "MainKt")
    }
    from(configurations.compileClasspath.get().map {
        if (it.isDirectory) it else zipTree(it)
    })
}
