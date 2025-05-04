package team.codemonsters.refactoringexample

import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication
import org.springframework.jms.annotation.EnableJms

@EnableJms
@SpringBootApplication
class RefactoringExampleApplication

fun main(args: Array<String>) {
	runApplication<RefactoringExampleApplication>(*args)
}
