package team.codemonsters.refactoringexample.service

import org.springframework.stereotype.Component
import team.codemonsters.refactoringexample.service.enum.PipelineType

@Component
class PipelineServiceSelector {
    fun choosePipelineService(messagePipeline: PipelineType): PipelineService {
        return when (messagePipeline) {
            PipelineType.PSBalance -> PipelineServicePS()
            PipelineType.StandardEnvelope -> PipelineServiceEnvelope()
        }
    }
}