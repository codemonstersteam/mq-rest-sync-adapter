package team.codemonsters.refactoringexample.service

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.jms.core.JmsTemplate
import org.springframework.test.context.ActiveProfiles
import team.codemonsters.refactoringexample.mq.BasicJmsRequest

@ActiveProfiles("test")
class MqSimpleTest(
    @Autowired val jmsTemplate: JmsTemplate,
    @Autowired val jmsPublisher: MqPublisher
) : AbstractIbmMqIntegrationTest() {

    @Test
    fun `should receive a message from MQ`() {
        // Arrange
        val message = "[]"
        val queue = "MY.TEST"
        // Act
        publishToQueue(queue, message)
        // Assert
        val response = jmsTemplate.receiveAndConvert(queue).toString()
        assertThat(response).isEqualTo(message)
    }

    private fun publishToQueue(queueName: String, messagePayload: String) {
        jmsPublisher.sendToMq(
            BasicJmsRequest(
                correlationId = "01",
                queue = queueName,
                payload = messagePayload,
                mapOf()
            ),
            originalMsgId = "01"
        )
    }

}