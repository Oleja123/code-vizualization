package codeVisualization.server.controller;

import codeVisualization.server.db.entity.User;
import codeVisualization.server.db.jpaRepository.UserRepository;
import codeVisualization.server.model.UsersDto;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpSession;
import org.junit.jupiter.api.*;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.*;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContext;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.crypto.password.PasswordEncoder;

import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

/**
 * Юнит-тесты AuthController.
 *
 * Мокируем только интерфейсы (UserRepository, PasswordEncoder, HttpServletRequest,
 * SecurityContext) — Mockito использует JDK proxy, Byte Buddy не нужен,
 * поэтому тесты работают на Java 23 без дополнительных флагов JVM.
 */
@ExtendWith(MockitoExtension.class)
@DisplayName("AuthController — юнит-тесты")
class AuthControllerTest {

    @Mock
    private UserRepository userRepository;

    @Mock
    private PasswordEncoder passwordEncoder;

    @Mock
    private HttpServletRequest httpRequest;

    @Mock
    private HttpSession httpSession;

    @Mock
    private SecurityContext securityContext;

    @InjectMocks
    private AuthController authController;

    @BeforeEach
    void setUp() {
        // Подменяем SecurityContextHolder перед каждым тестом
        SecurityContextHolder.setContext(securityContext);
    }

    @AfterEach
    void tearDown() {
        // Обязательно сбрасываем после теста, чтобы не влиять на соседние тесты
        SecurityContextHolder.clearContext();
    }

    // ──────────────────────────────────────────────────────────────
    //  POST /api/auth/login
    // ──────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("POST /api/auth/login")
    class Login {

        @Test
        @DisplayName("Успешный вход — 200 и установка сессии")
        void successfulLogin() {
            User dbUser = buildUser(1L, "alice", "$2a$hash");
            when(userRepository.findByName("alice")).thenReturn(Optional.of(dbUser));
            when(passwordEncoder.matches("secret", "$2a$hash")).thenReturn(true);
            when(httpRequest.getSession(true)).thenReturn(httpSession);
            // SecurityContext.setAuthentication() — void, stubbing не требуется

            ResponseEntity<?> response = authController.login(dto("alice", "secret"), httpRequest);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody()).asString().contains("Успешный");

            verify(securityContext).setAuthentication(argThat(auth ->
                    "alice".equals(auth.getPrincipal())
            ));
            verify(httpSession).setAttribute(
                    eq("SPRING_SECURITY_CONTEXT"),
                    any()
            );
        }

        @Test
        @DisplayName("Пользователь не найден — 401")
        void userNotFound() {
            when(userRepository.findByName("ghost")).thenReturn(Optional.empty());

            ResponseEntity<?> response = authController.login(dto("ghost", "any"), httpRequest);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
            assertThat(response.getBody()).asString().contains("Неверный");
            verify(passwordEncoder, never()).matches(any(), any());
        }

        @Test
        @DisplayName("Неверный пароль — 401")
        void wrongPassword() {
            User dbUser = buildUser(1L, "bob", "$2a$hash");
            when(userRepository.findByName("bob")).thenReturn(Optional.of(dbUser));
            when(passwordEncoder.matches("wrong", "$2a$hash")).thenReturn(false);

            ResponseEntity<?> response = authController.login(dto("bob", "wrong"), httpRequest);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
        }

        @Test
        @DisplayName("Успешный вход: сессия создаётся через getSession(true)")
        void sessionCreatedWithForceCreate() {
            User dbUser = buildUser(2L, "carol", "$2a$x");
            when(userRepository.findByName("carol")).thenReturn(Optional.of(dbUser));
            when(passwordEncoder.matches("pw", "$2a$x")).thenReturn(true);
            when(httpRequest.getSession(true)).thenReturn(httpSession);

            authController.login(dto("carol", "pw"), httpRequest);

            verify(httpRequest).getSession(true);
        }

        @Test
        @DisplayName("Поиск пользователя всегда выполняется по username из DTO")
        void searchByUsernameFromDto() {
            when(userRepository.findByName("dave")).thenReturn(Optional.empty());

            authController.login(dto("dave", "pw"), httpRequest);

            verify(userRepository).findByName("dave");
        }
    }

    // ──────────────────────────────────────────────────────────────
    //  POST /api/auth/logout
    // ──────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("POST /api/auth/logout")
    class Logout {

        @Test
        @DisplayName("Активная сессия инвалидируется — 200")
        void logoutWithActiveSession() {
            when(httpRequest.getSession(false)).thenReturn(httpSession);

            ResponseEntity<?> response = authController.logout(httpRequest);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(httpSession).invalidate();
        }

        @Test
        @DisplayName("Нет активной сессии — 200 (сессия просто отсутствует)")
        void logoutWithNoSession() {
            when(httpRequest.getSession(false)).thenReturn(null);

            ResponseEntity<?> response = authController.logout(httpRequest);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(httpSession, never()).invalidate();
        }

        @Test
        @DisplayName("SecurityContextHolder всегда очищается после logout")
        void securityContextClearedAfterLogout() {
            when(httpRequest.getSession(false)).thenReturn(null);

            authController.logout(httpRequest);

            // После вызова clearContext() контекст должен быть пустым
            assertThat(SecurityContextHolder.getContext().getAuthentication()).isNull();
        }
    }

    // ──────────────────────────────────────────────────────────────
    //  GET /api/auth/me
    // ──────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("GET /api/auth/me")
    class Me {

        @Test
        @DisplayName("Аутентифицированный пользователь — 200 и username")
        void authenticatedUser() {
            Authentication auth = authenticatedAs("alice");
            when(securityContext.getAuthentication()).thenReturn(auth);

            ResponseEntity<?> response = authController.currentUser(httpRequest);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody()).isEqualTo("alice");
        }

        @Test
        @DisplayName("Authentication == null — 401")
        void noAuthentication() {
            when(securityContext.getAuthentication()).thenReturn(null);

            ResponseEntity<?> response = authController.currentUser(httpRequest);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
        }

        @Test
        @DisplayName("Анонимный пользователь (principal = 'anonymousUser') — 401")
        void anonymousUser() {
            Authentication anon = mock(Authentication.class);
            when(anon.isAuthenticated()).thenReturn(true);
            when(anon.getPrincipal()).thenReturn("anonymousUser");
            when(securityContext.getAuthentication()).thenReturn(anon);

            ResponseEntity<?> response = authController.currentUser(httpRequest);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
        }

        @Test
        @DisplayName("Не аутентифицирован (isAuthenticated=false) — 401")
        void notAuthenticated() {
            Authentication auth = mock(Authentication.class);
            when(auth.isAuthenticated()).thenReturn(false);
            when(securityContext.getAuthentication()).thenReturn(auth);

            ResponseEntity<?> response = authController.currentUser(httpRequest);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
        }
    }

    // ──────────────────────────────────────────────────────────────
    //  POST /api/auth/register
    // ──────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("POST /api/auth/register")
    class Register {

        @Test
        @DisplayName("Новый пользователь — 201 Created")
        void registerNewUser() {
            when(userRepository.findByName("newbie")).thenReturn(Optional.empty());
            when(passwordEncoder.encode("pass123")).thenReturn("$2a$encoded");

            ResponseEntity<?> response =
                    authController.register(dto("newbie", "pass123"), httpRequest);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.CREATED);
            assertThat(response.getBody()).asString().contains("создан");
        }

        @Test
        @DisplayName("Пользователь с таким именем уже существует — 409 Conflict")
        void registerDuplicateUser() {
            when(userRepository.findByName("alice"))
                    .thenReturn(Optional.of(buildUser(1L, "alice", "$2a$x")));

            ResponseEntity<?> response =
                    authController.register(dto("alice", "pass"), httpRequest);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.CONFLICT);
            assertThat(response.getBody()).asString().contains("существует");
        }

        @Test
        @DisplayName("При успешной регистрации пароль хэшируется и пользователь сохраняется")
        void passwordEncodedAndUserSaved() {
            when(userRepository.findByName("dave")).thenReturn(Optional.empty());
            when(passwordEncoder.encode("rawpw")).thenReturn("$2a$hashed");

            authController.register(dto("dave", "rawpw"), httpRequest);

            ArgumentCaptor<User> captor = ArgumentCaptor.forClass(User.class);
            verify(userRepository).save(captor.capture());

            User saved = captor.getValue();
            assertThat(saved.getName()).isEqualTo("dave");
            assertThat(saved.getPassword()).isEqualTo("$2a$hashed");
        }

        @Test
        @DisplayName("При конфликте save() не вызывается")
        void noSaveOnConflict() {
            when(userRepository.findByName("alice"))
                    .thenReturn(Optional.of(buildUser(1L, "alice", "$2a$x")));

            authController.register(dto("alice", "pw"), httpRequest);

            verify(userRepository, never()).save(any());
        }

        @Test
        @DisplayName("Зарегистрированный пользователь не логинится автоматически — сессия не создаётся")
        void noAutoLoginAfterRegister() {
            when(userRepository.findByName("frank")).thenReturn(Optional.empty());
            when(passwordEncoder.encode(any())).thenReturn("$2a$x");

            authController.register(dto("frank", "pw"), httpRequest);

            verify(httpRequest, never()).getSession(anyBoolean());
        }
    }

    // ──────────────────────────────────────────────────────────────
    //  Вспомогательные методы
    // ──────────────────────────────────────────────────────────────

    private static UsersDto dto(String username, String rawPassword) {
        return new UsersDto(username, rawPassword);
    }

    private static User buildUser(Long id, String name, String encodedPassword) {
        User u = new User();
        u.setId(id);
        u.setName(name);
        u.setPassword(encodedPassword);
        return u;
    }

    /** Создаёт мок Authentication, притворяющийся аутентифицированным реальным пользователем. */
    private static Authentication authenticatedAs(String username) {
        Authentication auth = mock(Authentication.class);
        when(auth.isAuthenticated()).thenReturn(true);
        when(auth.getPrincipal()).thenReturn(username);
        when(auth.getName()).thenReturn(username);
        return auth;
    }
}