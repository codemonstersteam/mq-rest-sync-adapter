package team.codemonsters.refactoringexample.service

import com.ibm.mq.constants.CMQC
import com.ibm.msg.client.jakarta.jms.JmsConstants
import jakarta.jms.Message
import org.springframework.jms.core.JmsTemplate
import org.springframework.stereotype.Component
import org.springframework.util.StringUtils
import reactor.core.publisher.Mono
import team.codemonsters.refactoringexample.mq.BasicJmsRequest
import team.codemonsters.refactoringexample.mq.MqPublishedResponce


@Component
class MqPublisher(
    private val jmsTemplate: JmsTemplate
) {

    fun publishToMq(jmsRequest: BasicJmsRequest<String>, originalMsgId: String?): Mono<MqPublishedResponce<String>> =
        Mono.fromCallable {
            jmsTemplate.convertAndSend(
                jmsRequest.queue,
                jmsRequest.payload
            ) { message -> fillMessageWithHeaders(message, jmsRequest.headers, originalMsgId) }
        }
            .map {
                MqPublishedResponce(
                    correlationId = jmsRequest.correlationId,
                    status = "Success",
                    jmsRequest = jmsRequest
                )
            }

    fun sendToMq(jmsRequest: BasicJmsRequest<String>, originalMsgId: String?) {
        jmsTemplate.convertAndSend(
                jmsRequest.queue,
                jmsRequest.payload
            ) { message -> fillMessageWithHeaders(message, jmsRequest.headers, originalMsgId) }
    }

    private fun fillMessageWithHeaders(
        message: Message,
        headers: Map<String, String>,
        originalMsgId: String?
    ): Message {
        message.jmsCorrelationID = originalMsgId
        message.setIntProperty(JmsConstants.JMS_IBM_MSGTYPE, CMQC.MQMT_DATAGRAM)
        headers.forEach { (key: String, value: String) ->
            if (StringUtils.hasLength(key) && StringUtils.hasLength(value)) {
                message.setStringProperty(key, value)
            }
        }
        return message;
    }

}