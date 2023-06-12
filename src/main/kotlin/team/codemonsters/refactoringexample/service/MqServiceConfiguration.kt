package team.codemonsters.refactoringexample.service

import org.slf4j.LoggerFactory
import org.springframework.beans.factory.config.BeanDefinition
import org.springframework.beans.factory.config.ConfigurableListableBeanFactory
import org.springframework.beans.factory.support.BeanDefinitionBuilder
import org.springframework.beans.factory.support.BeanDefinitionRegistry
import org.springframework.beans.factory.support.BeanDefinitionRegistryPostProcessor
import org.springframework.boot.context.properties.bind.Bindable
import org.springframework.boot.context.properties.bind.Binder
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.core.env.ConfigurableEnvironment
import team.codemonsters.refactoringexample.configuration.MqConfiguration
import team.codemonsters.refactoringexample.configuration.MqServiceCfg

@Configuration
class MqServiceConfiguration {

    @Bean
    fun createMqServiceConfigurer(environment: ConfigurableEnvironment): BeanDefinitionRegistryPostProcessor {
        return MyPostProcessor(environment)
    }

    class MyPostProcessor(val environment: ConfigurableEnvironment): BeanDefinitionRegistryPostProcessor {

        private val log = LoggerFactory.getLogger(MyPostProcessor::class.java)!!

        override fun postProcessBeanFactory(beanFactory: ConfigurableListableBeanFactory) {
        }

        override fun postProcessBeanDefinitionRegistry(registry: BeanDefinitionRegistry) {
            val mqConfiguration = Binder.get(environment)
                .bind(MQ_PREFIX, Bindable.of(MqConfiguration::class.java))
                .orElseThrow { IllegalStateException() }

            mqConfiguration.configs.forEach { (k, v) ->
                log.debug("Adding MQ Service for key: ${k}")
                registry.registerBeanDefinition(getBeanName(k), buildBeanDefinition(v, MqHttpService::class.java))
            }
        }

        private fun buildBeanDefinition(config: MqServiceCfg, clazz: Class<out Any>): BeanDefinition {

            return BeanDefinitionBuilder.genericBeanDefinition(clazz).setLazyInit(true)
                .addConstructorArgReference("mqPublisher")
                .addConstructorArgValue(config)
                .addConstructorArgReference("restConfiguration")
                .addConstructorArgReference("jmsListenerContainerFactory")
                .addConstructorArgReference("pipelineServiceSelector")
                .beanDefinition;
        }

    }

    companion object{
        private const val MQ_SERVICE_PREFIX ="mq.http-service"
        private const val MQ_PREFIX = "mq"
        fun getBeanName(serviceName: String): String{
            return "$MQ_SERVICE_PREFIX.${serviceName}"
        }
    }

}