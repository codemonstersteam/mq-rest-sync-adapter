package team.codemonsters.mqrestsyncadapter.test;

import io.cucumber.junit.Cucumber;
import io.cucumber.junit.CucumberOptions;
import org.junit.runner.RunWith;

@RunWith(Cucumber.class)
@CucumberOptions(
    features = "src/test/resources/features",
    glue = "team.codemonsters.mqrestsyncadapter.test.stepdefs",
    plugin = {"pretty", "html:build/reports/cucumber/report.html"}
)
public class CucumberTestRunner {
} 