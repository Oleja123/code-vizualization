package e2e.pages;

import org.openqa.selenium.By;
import org.openqa.selenium.WebDriver;
import org.openqa.selenium.WebElement;
import org.openqa.selenium.support.ui.ExpectedConditions;
import org.openqa.selenium.support.ui.WebDriverWait;

import java.time.Duration;

/**
 * Page Object для экрана авторизации / регистрации.
 *
 * Vue-приложение рендерит единственный div#app. Пока пользователь не
 * авторизован, поверх него показывается .auth-overlay с формой входа
 * или регистрации. Этот класс инкапсулирует все взаимодействия с ней.
 *
 * Локаторы выбраны по типу элемента + атрибуту placeholder / тексту,
 * чтобы тест не ломался при изменении CSS-классов или порядка элементов.
 */
public class AuthPage {

    private final WebDriver driver;
    private final WebDriverWait wait;

    // ---- Локаторы ----

    /** Поле «Логин» (есть и в форме входа, и в форме регистрации). */
    private static final By USERNAME_INPUT =
            By.cssSelector("input[autocomplete='username'], input[placeholder='Имя пользователя']");

    /** Поле «Пароль» при входе. */
    private static final By PASSWORD_LOGIN_INPUT =
            By.cssSelector("input[autocomplete='current-password']");

    /** Поле «Пароль» при регистрации (другой autocomplete). */
    private static final By PASSWORD_REGISTER_INPUT =
            By.cssSelector("input[autocomplete='new-password']");

    /** Кнопка «Войти» / «Зарегистрироваться» — кнопка btn-primary. */
    private static final By SUBMIT_BUTTON =
            By.cssSelector(".btn-primary");

    /** Кнопка переключения режима (ссылка «Нет аккаунта? Зарегистрироваться» / «Уже есть аккаунт? Войти»). */
    private static final By TOGGLE_MODE_BUTTON =
            By.cssSelector(".link-btn");

    /** Блок с сообщением об ошибке. */
    private static final By ERROR_MESSAGE =
            By.cssSelector(".error");

    /** Заголовок формы (h2) — содержит «Вход в систему» или «Регистрация». */
    private static final By FORM_TITLE =
            By.cssSelector(".auth-card h2");

    /** Карточка авторизации — признак того, что форма отображается. */
    private static final By AUTH_CARD =
            By.cssSelector(".auth-card");

    public AuthPage(WebDriver driver) {
        this.driver = driver;
        this.wait = new WebDriverWait(driver, Duration.ofSeconds(15));
    }

    // =========================================================
    //  Проверки состояния
    // =========================================================

    /** Ждёт появления формы авторизации. */
    public AuthPage waitUntilVisible() {
        wait.until(ExpectedConditions.visibilityOfElementLocated(AUTH_CARD));
        return this;
    }

    /** Возвращает текущий заголовок формы. */
    public String getFormTitle() {
        return wait.until(ExpectedConditions.visibilityOfElementLocated(FORM_TITLE)).getText();
    }

    /** Возвращает текст ошибки (или пустую строку, если ошибки нет). */
    public String getErrorMessage() {
        try {
            WebElement err = driver.findElement(ERROR_MESSAGE);
            return err.isDisplayed() ? err.getText() : "";
        } catch (org.openqa.selenium.NoSuchElementException e) {
            return "";
        }
    }

    // =========================================================
    //  Навигация между режимами
    // =========================================================

    /** Переключает форму в режим регистрации. */
    public AuthPage switchToRegister() {
        WebElement toggle = wait.until(
                ExpectedConditions.elementToBeClickable(TOGGLE_MODE_BUTTON));
        if (!getFormTitle().contains("Регистрация")) {
            toggle.click();
            wait.until(ExpectedConditions.textToBe(FORM_TITLE, "Регистрация"));
        }
        return this;
    }

    /** Переключает форму в режим входа. */
    public AuthPage switchToLogin() {
        WebElement toggle = wait.until(
                ExpectedConditions.elementToBeClickable(TOGGLE_MODE_BUTTON));
        if (!getFormTitle().contains("Вход")) {
            toggle.click();
            wait.until(ExpectedConditions.textToBePresentInElementLocated(FORM_TITLE, "Вход"));
        }
        return this;
    }

    // =========================================================
    //  Действия
    // =========================================================

    /**
     * Заполняет форму регистрации и отправляет её.
     *
     * @param username логин нового пользователя
     * @param password пароль (минимум 6 символов)
     */
    public AuthPage register(String username, String password) {
        switchToRegister();
        fillUsername(username);
        fillPasswordInRegisterForm(password);
        clickSubmit();
        return this;
    }

    /**
     * Заполняет форму входа и отправляет её.
     *
     * @param username логин
     * @param password пароль
     */
    public AuthPage login(String username, String password) {
        switchToLogin();
        fillUsername(username);
        fillPasswordInLoginForm(password);
        clickSubmit();
        return this;
    }

    // =========================================================
    //  Приватные хелперы
    // =========================================================

    private void fillUsername(String username) {
        WebElement input = wait.until(ExpectedConditions.visibilityOfElementLocated(USERNAME_INPUT));
        input.clear();
        input.sendKeys(username);
    }

    private void fillPasswordInLoginForm(String password) {
        WebElement input = wait.until(
                ExpectedConditions.visibilityOfElementLocated(PASSWORD_LOGIN_INPUT));
        input.clear();
        input.sendKeys(password);
    }

    private void fillPasswordInRegisterForm(String password) {
        WebElement input = wait.until(
                ExpectedConditions.visibilityOfElementLocated(PASSWORD_REGISTER_INPUT));
        input.clear();
        input.sendKeys(password);
    }

    private void clickSubmit() {
        wait.until(ExpectedConditions.elementToBeClickable(SUBMIT_BUTTON)).click();
    }
}
