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

dependencies{
    implementation("com.google.code.gson:gson:2.8.2")
    implementation("org.hid4java:hid4java:0.7.0")
}


tasks.withType<KotlinCompile>() {
    kotlinOptions.jvmTarget = "1.8"
}

application {
    mainClass.set("MainKt")
}

tasks.jar {
    duplicatesStrategy = DuplicatesStrategy.INCLUDE
    archiveFileName.set("scale-tools.jar")
    destinationDirectory.set(file("../libs"))
    manifest.attributes.apply {
        put("Main-Class", "MainKt")
    }
    from(configurations.compileClasspath.get().map {
        if (it.isDirectory) it else zipTree(it)
    })
}
