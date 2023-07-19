package team.codemonsters.refactoringexample.service

import jakarta.annotation.PostConstruct
import jakarta.annotation.PreDestroy
import jakarta.jms.MessageListener
import org.slf4j.LoggerFactory
import org.springframework.jms.config.JmsListenerContainerFactory
import org.springframework.jms.config.SimpleJmsListenerEndpoint
import org.springframework.jms.listener.DefaultMessageListenerContainer
import org.springframework.jms.support.converter.MessagingMessageConverter
import org.springframework.messaging.Message
import org.springframework.util.ErrorHandler
import team.codemonsters.refactoringexample.client.RESTClient
import team.codemonsters.refactoringexample.configuration.MqServiceCfg
import team.codemonsters.refactoringexample.configuration.RestConfiguration


class MqHttpService(
    val mqPublisher: MqPublisher,
    val mqConfig: MqServiceCfg,
    val restConfiguration: RestConfiguration,
    val jmsFactory: JmsListenerContainerFactory<DefaultMessageListenerContainer>,
    val pipelineServiceSelector: PipelineServiceSelector,
    val restClient: RESTClient
) {
    private val log = LoggerFactory.getLogger(MqHttpService::class.java)!!

    private var container: DefaultMessageListenerContainer? = null
    private var messageConverter = MessagingMessageConverter()

    @PostConstruct
    fun registerJmsListener() {

        log.debug("Registering JMS listener for queue ${mqConfig.REQ}")
        val endpoint = SimpleJmsListenerEndpoint()
        endpoint.id = "myJmsEndpoint"
        endpoint.destination = mqConfig.REQ
        endpoint.messageListener = MessageListener { message ->
            log.info("Incoming transmission: ")
            log.info("Raw message: ${message}")
            //log.info("payload: ${message.payload}")
            log.info("Selected pipeline: ${mqConfig.messagePipeline}")

            pipelineServiceSelector.choosePipelineService(mqConfig.messagePipeline).receiveMessage(
                messageConverter.fromMessage(message) as Message<Any>,
                mqPublisher,
                mqConfig,
                restConfiguration,
                restClient
            )
        }
        container = jmsFactory.createListenerContainer(endpoint)
        container!!.errorHandler = ErrorHandler {
            it.printStackTrace()
        }
        container!!.initialize()
        container!!.start()
    }

    @PreDestroy
    fun shutdown() {
        if (container != null) {
            container!!.shutdown()
            container!!.destroy()
        }
    }


}