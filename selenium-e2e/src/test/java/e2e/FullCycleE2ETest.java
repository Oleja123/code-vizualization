package e2e;

import e2e.pages.AuthPage;
import e2e.pages.MainPage;
import org.junit.jupiter.api.*;
import org.openqa.selenium.By;

import static org.junit.jupiter.api.Assertions.*;

/**
 * E2E-тест: полный пользовательский цикл
 *
 * Сценарий и логика входов:
 *
 *  Шаг 1 — без сессии:  просто открываем страницу
 *  Шаг 2 — без сессии:  регистрируемся → сессия остаётся открытой
 *  Шаг 3 — своя сессия: тест выхода/входа — логинимся, выходим, логинимся снова
 *  Шаг 4 — продолжаем:  сессия от шага 3 ещё жива → просто переходим на вкладку
 *  Шаг 5 — продолжаем:  сессия от шага 4 ещё жива → просто переходим на вкладку
 *  Шаг 6 — своя сессия: тест выхода — логинимся, выходим
 *  Шаг 7 — своя сессия: тест что refresh не восстанавливает сессию
 *
 * Итого логинов: 4 (шаги 2→3→6→7) вместо прежних 6.
 * Браузер открывается один раз на весь прогон.
 */
@TestMethodOrder(MethodOrderer.OrderAnnotation.class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
@DisplayName("E2E: Полный цикл — регистрация → вход → схема → метрики → выход")
class FullCycleE2ETest extends BaseTest {

    private final String username = TestConfig.TEST_USERNAME;
    private final String password = TestConfig.TEST_PASSWORD;

    // =========================================================
    //  Шаг 1: Открытие приложения
    // =========================================================

    @Test
    @Order(1)
    @DisplayName("Шаг 1 — Открытие приложения показывает форму авторизации")
    void step1_openApp_showsAuthForm() {
        driver.get(TestConfig.APP_URL);

        String title = new AuthPage(driver).waitUntilVisible().getFormTitle();
        assertTrue(title.contains("Вход"),
                "Ожидался заголовок 'Вход в систему', но получено: " + title);
    }

    // =========================================================
    //  Шаг 2: Регистрация — сессия остаётся для шагов 4 и 5
    // =========================================================

    @Test
    @Order(2)
    @DisplayName("Шаг 2 — Регистрация нового пользователя")
    void step2_register_newUser() {
        driver.get(TestConfig.APP_URL);
        AuthPage authPage = new AuthPage(driver).waitUntilVisible();

        authPage.switchToRegister();
        assertEquals("Регистрация", authPage.getFormTitle());

        // После регистрации Vue автоматически логинит — сессия остаётся открытой
        authPage.register(username, password);
        MainPage mainPage = new MainPage(driver).waitUntilLoaded();

        assertEquals(username, mainPage.getLoggedInUsername(),
                "Имя пользователя в шапке должно совпадать с зарегистрированным");

        // Сессия НЕ сбрасывается — шаг 3 продолжит с ней
    }

    // =========================================================
    //  Шаг 3: Выход и повторный вход
    //  Сессия от шага 2 жива → выходим → логинимся снова
    //  После теста сессия остаётся для шагов 4 и 5
    // =========================================================

    @Test
    @Order(3)
    @DisplayName("Шаг 3 — Выход и повторный вход с теми же учётными данными")
    void step3_logout_and_login() {
        // Хедер уже виден — используем сессию от шага 2
        MainPage mainPage = new MainPage(driver).waitUntilLoaded();

        // Выходим
        AuthPage authPage = mainPage.logout();
        assertTrue(authPage.getFormTitle().contains("Вход"),
                "После выхода должна отображаться форма Вход в систему");

        // Входим снова — сессия остаётся для шагов 4 и 5
        authPage.login(username, password);
        mainPage = new MainPage(driver).waitUntilLoaded();

        assertEquals(username, mainPage.getLoggedInUsername(),
                "После повторного входа имя пользователя должно совпадать");
    }

    // =========================================================
    //  Шаг 4: Трассировка схемы
    //  Сессия от шага 3 жива — логин не нужен
    // =========================================================

    @Test
    @Order(4)
    @DisplayName("Шаг 4 — Трассировка схемы: генерация → трассировка → шаг → стоп")
    void step4_flowchartTracer() {
        // Сессия от шага 3 ещё жива — просто переходим на вкладку
        MainPage mainPage = new MainPage(driver).waitUntilLoaded();
        mainPage.goToTracer();

        MainPage.FlowchartPanel panel = mainPage.flowchartPanel();

        // 1. Генерируем блок-схему (код уже загружен — пример factorial)
        panel.generate();
        assertTrue(panel.isFlowchartVisible(),
                "SVG блок-схема должна отображаться после генерации");

        // 2. Кнопка «Начать трассировку» появилась
        assertTrue(panel.isStartTracingAvailable(),
                "Кнопка 'Начать трассировку' должна быть доступна после генерации");

        // 3. Запускаем трассировку и делаем два шага
        panel.startTracing();
        panel.stepForward();
        panel.stepForward();

        // 4. Останавливаем
        panel.stopTracing();
        assertTrue(panel.isStartTracingAvailable(),
                "После остановки кнопка 'Начать трассировку' должна снова появиться");
    }

    // =========================================================
    //  Шаг 5: Метрики
    //  Сессия от шага 4 жива — логин не нужен
    // =========================================================

    @Test
    @Order(5)
    @DisplayName("Шаг 5 — Метрики: запуск подсчёта для уже загруженного кода")
    void step5_metrics() {
        // Сессия от шага 4 ещё жива — просто переходим на вкладку
        MainPage mainPage = new MainPage(driver).waitUntilLoaded();
        mainPage.goToMetrics();

        MainPage.MetricsPanel panel = mainPage.metricsPanel();
        panel.clickCalculate();
        panel.waitForResults();

        assertTrue(panel.isSuccessMessageVisible(),
                "Должно отображаться сообщение '✓ Метрики подсчитаны'");
        assertTrue(panel.getFunctionCardCount() > 0,
                "Должна быть хотя бы одна карточка функции");
        assertFalse(panel.getFunctionCountBadge().isBlank(),
                "Бейдж с количеством функций не должен быть пустым");
    }

    // =========================================================
    //  Шаг 6: Выход из системы
    //  Сессия от шага 5 жива → выходим
    // =========================================================

    @Test
    @Order(6)
    @DisplayName("Шаг 6 — Выход: форма авторизации снова отображается")
    void step6_logout_showsAuthForm() {
        MainPage mainPage = new MainPage(driver).waitUntilLoaded();
        AuthPage authPage = mainPage.logout();

        assertTrue(authPage.getFormTitle().contains("Вход"),
                "После выхода должна отображаться форма 'Вход в систему'");
        assertTrue(
                driver.findElements(By.cssSelector(".app-header")).isEmpty(),
                "Хедер приложения не должен отображаться после выхода");
    }

    // =========================================================
    //  Шаг 7: Refresh не восстанавливает сессию
    //  Нет сессии после шага 6 — логинимся, выходим, refresh
    // =========================================================

    @Test
    @Order(7)
    @DisplayName("Шаг 7 — После выхода перезагрузка страницы не восстанавливает сессию")
    void step7_sessionNotRestoredAfterLogout() {
        // После шага 6 мы уже разлогинены — логинимся заново чтобы было что проверять
        driver.get(TestConfig.APP_URL);
        new AuthPage(driver).waitUntilVisible().login(username, password);
        new MainPage(driver).waitUntilLoaded().logout();

        driver.navigate().refresh();

        assertTrue(
                new AuthPage(driver).waitUntilVisible().getFormTitle().contains("Вход"),
                "После обновления страницы сессия не должна быть восстановлена");
    }
}
