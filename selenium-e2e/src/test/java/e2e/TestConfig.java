package e2e;

/**
 * Центральное хранилище настроек тестового окружения.
 *
 * Значения берутся из системных свойств JVM, что позволяет управлять
 * конфигурацией без перекомпиляции:
 *
 *   mvn test -Dapp.url=http://localhost:3000 -Dheadless=true
 *
 * Если свойство не задано — используется значение по умолчанию.
 */
public final class TestConfig {

    private TestConfig() {}

    /** Базовый URL фронтенда Vue-приложения. */
    public static final String APP_URL =
            System.getProperty("app.url", "http://localhost:3000");

    /** true — Chrome запускается без GUI (нужно в CI / Docker). */
    public static final boolean HEADLESS =
            Boolean.parseBoolean(System.getProperty("headless", "false"));

    /** Максимальное время ожидания элемента (секунды). */
    public static final int WAIT_SECONDS =
            Integer.parseInt(System.getProperty("wait.seconds", "15"));

    /** Пауза между действиями при «медленном» режиме (мс). */
    public static final int SLOW_DOWN_MS =
            Integer.parseInt(System.getProperty("slow.down.ms", "0"));

    // ---- Учётные данные для E2E-теста ----
    // Каждый запуск тестов генерирует уникальный логин, чтобы не зависеть
    // от состояния БД. Имя строится как "e2e_<timestamp>".
    public static final String TEST_USERNAME =
            System.getProperty("test.username", "e2e_" + System.currentTimeMillis());

    public static final String TEST_PASSWORD =
            System.getProperty("test.password", "Qwerty123!");

    // ---- Тестовый C-код для раздела «Трассировка кода» ----
    public static final String SAMPLE_C_CODE =
            "int main() {\n" +
            "  int x = 5;\n" +
            "  int y = 10;\n" +
            "  int sum = x + y;\n" +
            "  return sum;\n" +
            "}";

    // ---- Тестовый C-код для раздела «Метрики» ----
    public static final String METRICS_C_CODE =
            "#include <stdio.h>\n\n" +
            "int factorial(int n) {\n" +
            "  if (n <= 1) return 1;\n" +
            "  return n * factorial(n - 1);\n" +
            "}\n\n" +
            "int main() {\n" +
            "  int result = factorial(5);\n" +
            "  printf(\"%d\\n\", result);\n" +
            "  return 0;\n" +
            "}";
}
