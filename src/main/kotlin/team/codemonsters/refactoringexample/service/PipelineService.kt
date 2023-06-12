package team.codemonsters.refactoringexample.service

import org.springframework.messaging.Message
import team.codemonsters.refactoringexample.client.RESTClient
import team.codemonsters.refactoringexample.configuration.MqServiceCfg
import team.codemonsters.refactoringexample.configuration.RestConfiguration


interface PipelineService {
    fun receiveMessage(
        message: Message<Any>,
        mqPublisher: MqPublisher,
        mqConfig: MqServiceCfg,
        restConfiguration: RestConfiguration,
        restClient: RESTClient
    )
}