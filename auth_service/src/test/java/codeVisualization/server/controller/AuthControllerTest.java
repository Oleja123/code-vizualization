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
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.crypto.password.PasswordEncoder;

import java.util.Collections;
import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

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

    @InjectMocks
    private AuthController authController;

    @BeforeEach
    void setUp() {
        SecurityContextHolder.clearContext();
    }

    @AfterEach
    void tearDown() {
        SecurityContextHolder.clearContext();
    }

    // ──────────────────────────────────────────────────────────────
    //  POST /api/auth/login
    // ──────────────────────────────────────────────────────────────

    @Test
    void successfulLogin() {
        User user = buildUser(1L, "alice", "$2a$hash");
        when(userRepository.findByName("alice")).thenReturn(Optional.of(user));
        when(passwordEncoder.matches("secret", "$2a$hash")).thenReturn(true);
        when(httpRequest.getSession(true)).thenReturn(httpSession);

        ResponseEntity<?> response = authController.login(dto("alice", "secret"), httpRequest);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(response.getBody()).asString().contains("Успешный");

        Authentication auth = SecurityContextHolder.getContext().getAuthentication();
        assertThat(auth).isNotNull();
        assertThat(auth.getName()).isEqualTo("alice");

        verify(httpSession).setAttribute(eq("SPRING_SECURITY_CONTEXT"), any());
    }

    @Test
    void userNotFound() {
        when(userRepository.findByName("ghost")).thenReturn(Optional.empty());

        ResponseEntity<?> response = authController.login(dto("ghost", "pw"), httpRequest);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
        verify(passwordEncoder, never()).matches(any(), any());
    }

    @Test
    void wrongPassword() {
        User user = buildUser(1L, "bob", "$2a$hash");
        when(userRepository.findByName("bob")).thenReturn(Optional.of(user));
        when(passwordEncoder.matches("wrong", "$2a$hash")).thenReturn(false);

        ResponseEntity<?> response = authController.login(dto("bob", "wrong"), httpRequest);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
    }

    // ──────────────────────────────────────────────────────────────
    //  GET /api/auth/me
    // ──────────────────────────────────────────────────────────────

    @Test
    void authenticatedUser() {
        Authentication auth = new UsernamePasswordAuthenticationToken("alice", null, Collections.emptyList());
        SecurityContextHolder.getContext().setAuthentication(auth);

        ResponseEntity<?> response = authController.currentUser(httpRequest);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(response.getBody()).isEqualTo("alice");
    }

    @Test
    void noAuthentication() {
        SecurityContextHolder.getContext().setAuthentication(null);

        ResponseEntity<?> response = authController.currentUser(httpRequest);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
    }

    // ──────────────────────────────────────────────────────────────
    //  POST /api/auth/register
    // ──────────────────────────────────────────────────────────────

    @Test
    void registerNewUser() {
        when(userRepository.findByName("newbie")).thenReturn(Optional.empty());
        when(passwordEncoder.encode("pass123")).thenReturn("$2a$encoded");

        ResponseEntity<?> response = authController.register(dto("newbie", "pass123"), httpRequest);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.CREATED);
        assertThat(response.getBody()).asString().contains("создан");

        ArgumentCaptor<User> captor = ArgumentCaptor.forClass(User.class);
        verify(userRepository).save(captor.capture());
        assertThat(captor.getValue().getName()).isEqualTo("newbie");
    }

    @Test
    void registerDuplicateUser() {
        when(userRepository.findByName("alice")).thenReturn(Optional.of(buildUser(1L, "alice", "$2a$hash")));

        ResponseEntity<?> response = authController.register(dto("alice", "pw"), httpRequest);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.CONFLICT);
        verify(userRepository, never()).save(any());
    }

    // ──────────────────────────────────────────────────────────────
    //  Вспомогательные методы
    // ──────────────────────────────────────────────────────────────

    private static UsersDto dto(String username, String password) {
        return new UsersDto(username, password);
    }

    private static User buildUser(Long id, String name, String encodedPassword) {
        User u = new User();
        u.setId(id);
        u.setName(name);
        u.setPassword(encodedPassword);
        return u;
    }
}