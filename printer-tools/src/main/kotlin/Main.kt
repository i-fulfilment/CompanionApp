import com.google.gson.Gson
import javax.print.DocFlavor
import javax.print.PrintServiceLookup
import javax.print.attribute.AttributeSet
import javax.print.attribute.HashAttributeSet
import javax.print.attribute.standard.Media
import javax.print.attribute.standard.MediaTray
import javax.print.attribute.standard.PrinterName


fun main(args: Array<String>) {
    val printServices = PrintServiceLookup.lookupPrintServices(null, null)

    val printers = mutableListOf<Printer>()

    for (printerService in printServices) {
        val trays = mutableListOf<Tray>()

        val printName = printerService.name
        val aset: AttributeSet = HashAttributeSet()
        aset.add(PrinterName(printName, null))
        val services = PrintServiceLookup.lookupPrintServices(null, aset)
        for (i in services.indices) {
            val service = services[i]
            val flavor: DocFlavor = DocFlavor.SERVICE_FORMATTED.PRINTABLE
            val o = service.getSupportedAttributeValues(Media::class.java, flavor, null)
            if (o != null && o.javaClass.isArray) {
                for (media in o as Array<Media?>) {
                    if (media is MediaTray) {
                        trays.add(Tray(media.toString()))
                    }
                }
            }
        }

        printers.add(Printer(printerService.name, trays))
    }
    val gson = Gson()
    print(gson.toJson(printers))
}
