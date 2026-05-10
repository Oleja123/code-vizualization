package e2e.pages;

import org.openqa.selenium.By;
import org.openqa.selenium.JavascriptExecutor;
import org.openqa.selenium.WebDriver;
import org.openqa.selenium.WebElement;
import org.openqa.selenium.support.ui.ExpectedConditions;
import org.openqa.selenium.support.ui.WebDriverWait;

import java.time.Duration;

/**
 * Page Object для главного экрана приложения (после авторизации).
 *
 * Покрывает:
 *  - хедер с навигацией (вкладки: Трассировка схемы / Трассировка кода /
 *    Анализ кода / Метрики)
 *  - отображение имени пользователя
 *  - кнопку «Выйти»
 *
 * Каждый раздел вынесен в отдельный внутренний класс-«панель»
 * (CodeTracerPanel, MetricsPanel), чтобы не смешивать локаторы.
 */
public class MainPage {

    protected final WebDriver driver;
    protected final WebDriverWait wait;

    // ---- Навигация ----
    private static final By NAV_TRACER =
            By.xpath("//button[contains(., 'Трассировка схемы')]");
    private static final By NAV_VISUALIZATION =
            By.xpath("//button[contains(., 'Трассировка кода')]");
    private static final By NAV_ANALYSIS =
            By.xpath("//button[contains(., 'Анализ кода')]");
    private static final By NAV_METRICS =
            By.xpath("//button[contains(., 'Метрики')]");

    // ---- Хедер ----
    private static final By USERNAME_LABEL =
            By.cssSelector(".username");
    private static final By LOGOUT_BUTTON =
            By.cssSelector(".logout-btn");
    private static final By APP_HEADER =
            By.cssSelector(".app-header");

    public MainPage(WebDriver driver) {
        this.driver = driver;
        this.wait = new WebDriverWait(driver, Duration.ofSeconds(15));
    }

    // =========================================================
    //  Ожидание загрузки
    // =========================================================

    /** Ждёт появления заголовка приложения — признак успешного входа. */
    public MainPage waitUntilLoaded() {
        wait.until(ExpectedConditions.visibilityOfElementLocated(APP_HEADER));
        return this;
    }

    // =========================================================
    //  Информация о пользователе
    // =========================================================

    /** Возвращает логин, отображаемый в шапке (без эмодзи 👤). */
    public String getLoggedInUsername() {
        String raw = wait.until(
                ExpectedConditions.visibilityOfElementLocated(USERNAME_LABEL)).getText();
        // Убираем возможный символ 👤 и пробелы
        return raw.replace("👤", "").trim();
    }

    // =========================================================
    //  Навигация между разделами
    // =========================================================

    public MainPage goToTracer() {
        waitAndClick(NAV_TRACER);
        return this;
    }

    public MainPage goToVisualization() {
        waitAndClick(NAV_VISUALIZATION);
        return this;
    }

    public MainPage goToAnalysis() {
        waitAndClick(NAV_ANALYSIS);
        return this;
    }

    public MainPage goToMetrics() {
        waitAndClick(NAV_METRICS);
        return this;
    }

    // =========================================================
    //  Выход
    // =========================================================

    /**
     * Кликает «Выйти» и возвращает AuthPage.
     * Вызов предполагает, что после клика откроется форма авторизации.
     */
    public AuthPage logout() {
        waitAndClick(LOGOUT_BUTTON);
        return new AuthPage(driver).waitUntilVisible();
    }

    // =========================================================
    //  Вложенные Page Object для конкретных разделов
    // =========================================================

    /**
     * Возвращает Panel для работы с разделом «Трассировка схемы».
     * Нужно предварительно вызвать goToTracer().
     */
    public FlowchartPanel flowchartPanel() {
        return new FlowchartPanel(driver, wait);
    }

    /**
     * Возвращает Panel для работы с разделом «Метрики».
     * Нужно предварительно вызвать goToMetrics().
     */
    public MetricsPanel metricsPanel() {
        return new MetricsPanel(driver, wait);
    }

    // =========================================================
    //  Внутренний хелпер
    // =========================================================

    protected WebElement waitAndClick(By locator) {
        WebElement el = wait.until(ExpectedConditions.elementToBeClickable(locator));
        el.click();
        return el;
    }

    // =========================================================
    //  Вложенный класс: панель «Трассировка схемы» (FlowchartTracer)
    // =========================================================

    public static class FlowchartPanel {

        private final WebDriver driver;
        private final WebDriverWait wait;

        // Кнопка «Сгенерировать схему» — видна когда phase === 'idle'
        private static final By GENERATE_BUTTON =
                By.cssSelector(".btn-generate");
        // Кнопка «⚡ Начать трассировку» — видна когда phase === 'ready'
        private static final By START_TRACE_BUTTON =
                By.cssSelector(".btn-trace");
        // Кнопка «Вперёд →» — ищем по тексту, т.к. обе кнопки имеют класс btn-step
        private static final By STEP_FORWARD_BUTTON =
                By.xpath("//button[contains(@class,'btn-step') and contains(text(),'Вперёд')]");
        // Кнопка «■ Стоп»
        private static final By STOP_BUTTON =
                By.cssSelector(".btn-stop");
        // SVG блок-схема — признак успешной генерации
        private static final By FLOWCHART_SVG =
                By.cssSelector("svg");
        // Панель переменных / состояния программы (появляется при трассировке)
        private static final By VARS_PANEL =
                By.cssSelector(".vars-toggle-btn, .vars-panel");

        FlowchartPanel(WebDriver driver, WebDriverWait wait) {
            this.driver = driver;
            this.wait = wait;
        }

        /**
         * Нажимает «Сгенерировать схему» и ждёт появления SVG.
         * Бэкенд flowchart-visualizer может отвечать несколько секунд.
         */
        public FlowchartPanel generate() {
            wait.until(ExpectedConditions.elementToBeClickable(GENERATE_BUTTON)).click();
            // Ждём SVG с увеличенным таймаутом
            new WebDriverWait(driver, Duration.ofSeconds(30))
                    .until(ExpectedConditions.visibilityOfElementLocated(FLOWCHART_SVG));
            return this;
        }

        /**
         * Нажимает «Начать трассировку».
         * Кнопка появляется только после успешной генерации схемы.
         */
        public FlowchartPanel startTracing() {
            wait.until(ExpectedConditions.elementToBeClickable(START_TRACE_BUTTON)).click();
            return this;
        }

        /**
         * Делает один шаг вперёд по трассировке.
         * Ждёт пока кнопка не станет активной (не disabled) — после startTracing()
         * интерпретатор загружает первый шаг асинхронно.
         */
        public FlowchartPanel stepForward() {
            // Сначала ждём появления кнопки
            WebElement btn = wait.until(
                    ExpectedConditions.presenceOfElementLocated(STEP_FORWARD_BUTTON));
            // Затем ждём пока она станет кликабельной (disabled снимается после загрузки шага)
            wait.until(ExpectedConditions.elementToBeClickable(STEP_FORWARD_BUTTON));
            btn.click();
            // Небольшая пауза чтобы интерпретатор обработал шаг перед следующим кликом
            try { Thread.sleep(300); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
            return this;
        }

        /**
         * Останавливает трассировку.
         */
        public FlowchartPanel stopTracing() {
            wait.until(ExpectedConditions.elementToBeClickable(STOP_BUTTON)).click();
            return this;
        }

        /** true, если SVG блок-схема отображается на экране. */
        public boolean isFlowchartVisible() {
            try {
                return driver.findElement(FLOWCHART_SVG).isDisplayed();
            } catch (org.openqa.selenium.NoSuchElementException e) {
                return false;
            }
        }

        /** true, если кнопка «Начать трассировку» доступна (схема сгенерирована). */
        public boolean isStartTracingAvailable() {
            try {
                return driver.findElement(START_TRACE_BUTTON).isDisplayed();
            } catch (org.openqa.selenium.NoSuchElementException e) {
                return false;
            }
        }
    }

    // =========================================================
    //  Вложенный класс: панель «Метрики»
    // =========================================================

    public static class MetricsPanel {

        private final WebDriver driver;
        private final WebDriverWait wait;

        // Textarea редактора в MetricsView
        private static final By CODE_TEXTAREA =
                By.cssSelector(".metrics-root .code-ta");
        // Кнопка «▶ Подсчитать метрики»
        private static final By CALCULATE_BUTTON =
                By.cssSelector(".btn-calc");
        // Сообщение «✓ Метрики подсчитаны»
        private static final By SUCCESS_MSG =
                By.cssSelector(".msg.ok");
        // Бейджи с количеством функций и переменных
        private static final By BADGE_FUNC_COUNT =
                By.xpath("//span[contains(@class,'badge') and contains(.,'Функций')]");
        private static final By BADGE_GLOBAL_VARS =
                By.xpath("//span[contains(@class,'badge') and contains(.,'Глоб.')]");
        // Карточки функций
        private static final By FN_CARDS =
                By.cssSelector(".fn-card");
        // Спиннер загрузки
        private static final By SPINNER =
                By.cssSelector(".spinner");

        MetricsPanel(WebDriver driver, WebDriverWait wait) {
            this.driver = driver;
            this.wait = wait;
        }

        /**
         * Вводит C-код в textarea раздела метрик.
         *
         * Используем click + Ctrl+A + sendKeys вместо прямого JS-присвоения value,
         * потому что Vue отслеживает нативные события клавиатуры через v-model,
         * а не изменение .value напрямую. sendKeys корректно триггерит все нужные
         * DOM-события (keydown / input / keyup).
         */
        public MetricsPanel enterCode(String code) {
            WebElement ta = wait.until(
                    ExpectedConditions.elementToBeClickable(CODE_TEXTAREA));
            ta.click();
            ta.sendKeys(org.openqa.selenium.Keys.chord(
                    org.openqa.selenium.Keys.CONTROL, "a"));
            ta.sendKeys(code);
            return this;
        }

        /** Нажимает кнопку «Подсчитать метрики». */
        public MetricsPanel clickCalculate() {
            wait.until(ExpectedConditions.elementToBeClickable(CALCULATE_BUTTON)).click();
            return this;
        }

        /**
         * Ждёт исчезновения спиннера и появления сообщения об успехе.
         * Таймаут увеличен до 30 с, т.к. бэкенд метрик может отвечать дольше.
         */
        public MetricsPanel waitForResults() {
            WebDriverWait longWait = new WebDriverWait(driver, Duration.ofSeconds(60));
            // Сначала ждём, пока спиннер исчезнет
            try {
                longWait.until(ExpectedConditions.invisibilityOfElementLocated(SPINNER));
            } catch (Exception ignored) {
                // Спиннер мог и не появиться, если ответ пришёл мгновенно
            }
            // Затем ждём зелёного сообщения об успехе
            longWait.until(ExpectedConditions.visibilityOfElementLocated(SUCCESS_MSG));
            return this;
        }

        /** true, если сообщение «✓ Метрики подсчитаны» отображается. */
        public boolean isSuccessMessageVisible() {
            try {
                return driver.findElement(SUCCESS_MSG).isDisplayed();
            } catch (org.openqa.selenium.NoSuchElementException e) {
                return false;
            }
        }

        /** Возвращает текст бейджа с количеством функций. */
        public String getFunctionCountBadge() {
            return wait.until(
                    ExpectedConditions.visibilityOfElementLocated(BADGE_FUNC_COUNT)).getText();
        }

        /** Возвращает количество карточек функций на экране. */
        public int getFunctionCardCount() {
            return driver.findElements(FN_CARDS).size();
        }
    }
}
