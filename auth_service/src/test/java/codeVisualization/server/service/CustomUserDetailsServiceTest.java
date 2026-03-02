package codeVisualization.server.service;

import codeVisualization.server.db.entity.User;
import codeVisualization.server.db.jpaRepository.UserRepository;
import org.junit.jupiter.api.*;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.*;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.security.core.userdetails.UsernameNotFoundException;

import java.util.Optional;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.Mockito.*;

/**
 * Юнит-тесты CustomUserDetailsService.
 *
 * Мокируем интерфейс UserRepository — JDK proxy, Java 23 совместимо.
 */
@ExtendWith(MockitoExtension.class)
@DisplayName("CustomUserDetailsService — юнит-тесты")
class CustomUserDetailsServiceTest {

    @Mock
    private UserRepository userRepository;

    @InjectMocks
    private CustomUserDetailsService userDetailsService;

    // ──────────────────────────────────────────────────────────────
    //  loadUserByUsername — успешный случай
    // ──────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("loadUserByUsername — пользователь найден")
    class UserFound {

        @Test
        @DisplayName("Возвращает UserDetails с правильным именем")
        void returnsCorrectUsername() {
            when(userRepository.findByName("alice")).thenReturn(Optional.of(user("alice", "$2a$hash")));

            UserDetails details = userDetailsService.loadUserByUsername("alice");

            assertThat(details.getUsername()).isEqualTo("alice");
        }

        @Test
        @DisplayName("Пароль в UserDetails совпадает с хэшем из БД")
        void passwordIsPreserved() {
            when(userRepository.findByName("bob")).thenReturn(Optional.of(user("bob", "$2a$secret")));

            UserDetails details = userDetailsService.loadUserByUsername("bob");

            assertThat(details.getPassword()).isEqualTo("$2a$secret");
        }

        @Test
        @DisplayName("UserDetails содержит роль ROLE_USER")
        void hasRoleUser() {
            when(userRepository.findByName("carol")).thenReturn(Optional.of(user("carol", "$2a$x")));

            UserDetails details = userDetailsService.loadUserByUsername("carol");

            assertThat(details.getAuthorities())
                    .extracting(Object::toString)
                    .containsExactly("ROLE_USER");
        }

        @Test
        @DisplayName("Аккаунт не заблокирован и не просрочен")
        void accountIsActive() {
            when(userRepository.findByName("dave")).thenReturn(Optional.of(user("dave", "$2a$x")));

            UserDetails details = userDetailsService.loadUserByUsername("dave");

            assertThat(details.isEnabled()).isTrue();
            assertThat(details.isAccountNonLocked()).isTrue();
            assertThat(details.isAccountNonExpired()).isTrue();
            assertThat(details.isCredentialsNonExpired()).isTrue();
        }

        @Test
        @DisplayName("findByName вызывается ровно один раз с переданным username")
        void repositoryCalledOnce() {
            when(userRepository.findByName("eve")).thenReturn(Optional.of(user("eve", "$2a$y")));

            userDetailsService.loadUserByUsername("eve");

            verify(userRepository, times(1)).findByName("eve");
        }
    }

    // ──────────────────────────────────────────────────────────────
    //  loadUserByUsername — пользователь не найден
    // ──────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("loadUserByUsername — пользователь не найден")
    class UserNotFound {

        @Test
        @DisplayName("Выбрасывает UsernameNotFoundException")
        void throwsException() {
            when(userRepository.findByName("ghost")).thenReturn(Optional.empty());

            assertThatThrownBy(() -> userDetailsService.loadUserByUsername("ghost"))
                    .isInstanceOf(UsernameNotFoundException.class);
        }

        @Test
        @DisplayName("Сообщение исключения содержит 'не найден'")
        void exceptionMessageMentionsNotFound() {
            when(userRepository.findByName("nobody")).thenReturn(Optional.empty());

            assertThatThrownBy(() -> userDetailsService.loadUserByUsername("nobody"))
                    .isInstanceOf(UsernameNotFoundException.class)
                    .hasMessageContaining("не найден");
        }

        @Test
        @DisplayName("При отсутствии пользователя репозиторий всё равно вызывается")
        void repositoryStillCalled() {
            when(userRepository.findByName("absent")).thenReturn(Optional.empty());

            assertThatThrownBy(() -> userDetailsService.loadUserByUsername("absent"))
                    .isInstanceOf(UsernameNotFoundException.class);

            verify(userRepository).findByName("absent");
        }
    }

    // ──────────────────────────────────────────────────────────────
    //  Вспомогательный метод
    // ──────────────────────────────────────────────────────────────

    private static User user(String name, String encodedPassword) {
        User u = new User();
        u.setId(1L);
        u.setName(name);
        u.setPassword(encodedPassword);
        return u;
    }
}