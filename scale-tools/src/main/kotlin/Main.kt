import org.hid4java.*
import kotlin.math.pow
import kotlin.system.exitProcess
import com.google.gson.Gson


const val VENDOR_ID = 0x0922
const val PRODUCT_ID = 0x8003
const val DATA_MODE_GRAMS = 2

fun main(args: Array<String>) {

    val hidServices = HidManager.getHidServices()
    val scales = hidServices.getHidDevice(VENDOR_ID, PRODUCT_ID, null)

    val gson = Gson()

    if(scales == null) {
        print(gson.toJson(Result(error = "Failed to connect to USB scales.", weight = 0)))
        return
    }

    try {
        val grams = getWeightInGrams(scales)
        print(gson.toJson(Result(error = null, weight = grams)))
        hidServices.shutdown()
        return
    } catch (e: Exception) {
        hidServices.shutdown()
        print(gson.toJson(Result(error = e.message, weight = 0)))
        exitProcess(1)
    }
}

fun getWeightInGrams(scales: HidDevice): Int {

    // Byte 0 == Report ID?
    // Byte 1 == Scale Status (1 == Fault, 2 == Stable @ 0, 3 == In Motion, 4 == Stable, 5 == Under 0, 6 == Over Weight, 7 == Requires Calibration, 8 == Requires Re-Zeroing)
    // Byte 2 == Weight Unit
    // Byte 3 == Data Scaling (decimal placement)
    // Byte 4 == Weight LSB
    // Byte 5 == Weight MSB
    val message = ByteArray(6)

    if(!scales.isOpen) {
        scales.open()

        if(!scales.isOpen) {
            throw java.lang.Exception("Could not connect to the scales.")
        }
    }

    scales.read(message, 1000)
    scales.close()

    if(message[1].toInt() != 4) {
        throw java.lang.Exception("Could not get a positive stable weight.")
    }

    val multiplied: Int = 256 * message[5].toUByte().toInt()

    val weight = (message[4].toUByte().toInt() + multiplied) * 10.0.pow(message[3].toDouble())

    if(message[2].toInt() == DATA_MODE_GRAMS){
        return weight.toInt()
    }

    val converted = weight * 28.3495231
    return converted.toInt()
}
