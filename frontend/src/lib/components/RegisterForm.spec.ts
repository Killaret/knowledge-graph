import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/svelte';
import RegisterForm from './RegisterForm.svelte';

const mockRegister = vi.fn();

vi.mock('$lib/stores/auth.svelte.js', () => ({
  register: (...args: any[]) => mockRegister(...args),
  isLoading: () => false,
  error: () => null
}));

vi.mock('$app/navigation', () => ({
  goto: vi.fn()
}));

describe('RegisterForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    cleanup();
  });

  it('renders registration form with all fields', () => {
    render(RegisterForm);

    expect(screen.getByLabelText(/логин/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^пароль/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/подтвердите пароль/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /зарегистрироваться/i })).toBeInTheDocument();
  });

  it('updates all form fields', async () => {
    render(RegisterForm);

    const loginInput = screen.getByLabelText(/логин/i);
    const emailInput = screen.getByLabelText(/email/i);
    const passwordInput = screen.getByLabelText(/^пароль/i);
    const confirmInput = screen.getByLabelText(/подтвердите пароль/i);

    await fireEvent.input(loginInput, { target: { value: 'newuser' } });
    await fireEvent.input(emailInput, { target: { value: 'newuser@example.com' } });
    await fireEvent.input(passwordInput, { target: { value: 'Password123!' } });
    await fireEvent.input(confirmInput, { target: { value: 'Password123!' } });

    expect(loginInput).toHaveValue('newuser');
    expect(emailInput).toHaveValue('newuser@example.com');
    expect(passwordInput).toHaveValue('Password123!');
    expect(confirmInput).toHaveValue('Password123!');
  });

  it('shows password requirements when password is entered', async () => {
    render(RegisterForm);

    const passwordInput = screen.getByLabelText(/^пароль/i);
    await fireEvent.input(passwordInput, { target: { value: 'weak' } });

    expect(screen.getByText(/требования к паролю/i)).toBeInTheDocument();
    expect(screen.getByText(/минимум 10 символов/i)).toBeInTheDocument();
    expect(screen.getByText(/заглавная буква/i)).toBeInTheDocument();
  });

  it('shows error when login is empty', async () => {
    const { container } = render(RegisterForm);

    const form = container.querySelector('form');
    await fireEvent.submit(form!);

    expect(screen.getByRole('alert')).toHaveTextContent(/введите логин/i);
  });

  it('shows error when passwords do not match', async () => {
    render(RegisterForm);

    const passwordInput = screen.getByLabelText(/^пароль/i);
    const confirmInput = screen.getByLabelText(/подтвердите пароль/i);

    await fireEvent.input(passwordInput, { target: { value: 'Password123!' } });
    await fireEvent.input(confirmInput, { target: { value: 'Different123!' } });

    expect(screen.getByText(/пароли не совпадают/i)).toBeInTheDocument();
  });

  it('submits registration with valid data', async () => {
    mockRegister.mockResolvedValue(true);
    render(RegisterForm);

    const loginInput = screen.getByLabelText(/логин/i);
    const emailInput = screen.getByLabelText(/email/i);
    const passwordInput = screen.getByLabelText(/^пароль/i);
    const confirmInput = screen.getByLabelText(/подтвердите пароль/i);
    const submitButton = screen.getByRole('button', { name: /зарегистрироваться/i });

    await fireEvent.input(loginInput, { target: { value: 'newuser' } });
    await fireEvent.input(emailInput, { target: { value: 'newuser@example.com' } });
    await fireEvent.input(passwordInput, { target: { value: 'Password123!' } });
    await fireEvent.input(confirmInput, { target: { value: 'Password123!' } });
    await fireEvent.click(submitButton);

    expect(mockRegister).toHaveBeenCalledWith('newuser', 'Password123!', 'newuser@example.com');
  });

  it('has link to login page', () => {
    render(RegisterForm);

    const loginLink = screen.getByRole('link', { name: /войти/i });
    expect(loginLink).toHaveAttribute('href', '/auth/login');
  });
});
