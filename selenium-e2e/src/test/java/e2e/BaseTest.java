package e2e;

import io.github.bonigarcia.wdm.WebDriverManager;
import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.TestInstance;
import org.openqa.selenium.By;
import org.openqa.selenium.JavascriptExecutor;
import org.openqa.selenium.WebDriver;
import org.openqa.selenium.WebElement;
import org.openqa.selenium.chrome.ChromeDriver;
import org.openqa.selenium.chrome.ChromeOptions;
import org.openqa.selenium.support.ui.ExpectedConditions;
import org.openqa.selenium.support.ui.WebDriverWait;

import java.time.Duration;

/**
 * Базовый класс для всех E2E-тестов.
 *
 * Браузер открывается ОДИН РАЗ на весь тестовый класс (@BeforeAll)
 * и закрывается после последнего теста (@AfterAll).
 * Между тестами очищаются только cookies (@BeforeEach) — окно не мелькает.
 *
 * Требует @TestInstance(PER_CLASS) на подклассе, чтобы @BeforeAll/@AfterAll
 * могли быть нестатическими и иметь доступ к полям экземпляра.
 */
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
public abstract class BaseTest {

    protected WebDriver driver;
    protected WebDriverWait wait;

    // =========================================================
    //  Lifecycle: один браузер на весь класс
    // =========================================================

    @BeforeAll
    void setupBrowser() {
        WebDriverManager.chromedriver().setup();
        driver = new ChromeDriver(buildChromeOptions());
        driver.manage().window().maximize();
        wait = new WebDriverWait(driver, Duration.ofSeconds(TestConfig.WAIT_SECONDS));
    }

    @AfterAll
    void tearDownBrowser() {
        if (driver != null) {
            driver.quit();
            driver = null;
        }
    }

    // Cookies намеренно НЕ сбрасываются между тестами —
    // шаги 4 и 5 переиспользуют сессию от предыдущих шагов.
    // Каждый шаг сам управляет своим состоянием входа/выхода.

    // =========================================================
    //  Фабрика ChromeOptions
    // =========================================================

    private ChromeOptions buildChromeOptions() {
        ChromeOptions options = new ChromeOptions();

        if (TestConfig.HEADLESS) {
            options.addArguments("--headless=new");
        }

        options.addArguments(
                "--no-sandbox",
                "--disable-dev-shm-usage",
                "--disable-gpu",
                "--window-size=1920,1080",
                "--disable-notifications",
                "--disable-infobars"
        );

        return options;
    }

    // =========================================================
    //  Вспомогательные методы
    // =========================================================

    protected WebElement waitAndClick(By locator) {
        WebElement el = wait.until(ExpectedConditions.elementToBeClickable(locator));
        slowDown();
        el.click();
        return el;
    }

    protected WebElement waitVisible(By locator) {
        return wait.until(ExpectedConditions.visibilityOfElementLocated(locator));
    }

    protected void clearAndType(By locator, String text) {
        WebElement el = waitVisible(locator);
        el.clear();
        el.sendKeys(text);
    }

    /**
     * Устанавливает значение textarea через JavaScript и диспатчит
     * события input + change — нужно для корректной реакции Vue.
     */
    protected void setTextareaValue(By locator, String text) {
        WebElement el = waitVisible(locator);
        JavascriptExecutor js = (JavascriptExecutor) driver;
        js.executeScript("arguments[0].value = arguments[1];", el, text);
        js.executeScript(
                "arguments[0].dispatchEvent(new Event('input', {bubbles:true}));" +
                "arguments[0].dispatchEvent(new Event('change', {bubbles:true}));",
                el);
    }

    protected boolean isTextPresent(String text) {
        return driver.getPageSource().contains(text);
    }

    protected void slowDown() {
        if (TestConfig.SLOW_DOWN_MS > 0) {
            try { Thread.sleep(TestConfig.SLOW_DOWN_MS); }
            catch (InterruptedException e) { Thread.currentThread().interrupt(); }
        }
    }

    protected void sleep(long ms) {
        try { Thread.sleep(ms); }
        catch (InterruptedException e) { Thread.currentThread().interrupt(); }
    }
}
