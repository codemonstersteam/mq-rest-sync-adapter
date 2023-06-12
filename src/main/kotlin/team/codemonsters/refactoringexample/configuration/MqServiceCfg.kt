package team.codemonsters.refactoringexample.configuration

import team.codemonsters.refactoringexample.service.enum.PipelineType
import java.time.LocalDateTime
import java.time.ZoneOffset
import java.time.format.DateTimeFormatter

data class MqServiceCfg(
    var REQ: String = "",
    var RES: String = "",
    var RECIEPT: String ="",
    var timeout: Long = 60000,
    var correlationID: String = "",
    var messagePipeline: PipelineType = PipelineType.StandardEnvelope,
    var createDateTime: String = "",
    var formatCreateDateTime: String = "",
    var values: Map<String, String> = mutableMapOf(),
    var type: MqServiceType = MqServiceType.ENVELOPE
) {
    fun getFormattedCreateDateTime(): String {
        val date = LocalDateTime.now().atZone(ZoneOffset.UTC)
        val formatter = DateTimeFormatter.ofPattern(formatCreateDateTime)
        return date.format(formatter)
    }
}