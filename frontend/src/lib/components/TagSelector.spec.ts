import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import TagSelector from './TagSelector.svelte';

describe('TagSelector', () => {
  const defaultTags = ['javascript', 'svelte', 'testing'];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders with existing tags', () => {
    render(TagSelector, { 
      props: { tags: defaultTags } 
    });

    expect(screen.getByText('#javascript')).toBeInTheDocument();
    expect(screen.getByText('#svelte')).toBeInTheDocument();
    expect(screen.getByText('#testing')).toBeInTheDocument();
  });

  it('renders empty state with placeholder', () => {
    render(TagSelector, { 
      props: { tags: [], placeholder: 'Add tags...' } 
    });

    expect(screen.getByPlaceholderText('Add tags...')).toBeInTheDocument();
  });

  it('adds new tag on Enter key', async () => {
    const onChange = vi.fn();
    
    render(TagSelector, { 
      props: { tags: [], onChange } 
    });

    const input = screen.getByLabelText(/add new tag/i);
    await userEvent.type(input, 'newtag');
    await userEvent.keyboard('{Enter}');

    expect(onChange).toHaveBeenCalledWith(['newtag']);
  });

  it('adds tag when trimmable spaces are present', async () => {
    const onChange = vi.fn();
    
    render(TagSelector, { 
      props: { tags: [], onChange } 
    });

    const input = screen.getByLabelText(/add new tag/i);
    await userEvent.type(input, '  spaced-tag  ');
    await userEvent.keyboard('{Enter}');

    // Тег должен быть добавлен в нижнем регистре и без пробелов
    expect(onChange).toHaveBeenCalledWith(['spaced-tag']);
  });

  it('removes tag when clicking remove button', async () => {
    const onChange = vi.fn();
    
    render(TagSelector, { 
      props: { tags: defaultTags, onChange } 
    });

    const removeButton = screen.getByLabelText(/remove tag javascript/i);
    await userEvent.click(removeButton);

    expect(onChange).toHaveBeenCalledWith(['svelte', 'testing']);
  });

  it('removes last tag on Backspace when input is empty', async () => {
    const onChange = vi.fn();
    
    render(TagSelector, { 
      props: { tags: defaultTags, onChange } 
    });

    const input = screen.getByLabelText(/add new tag/i);
    await userEvent.click(input);
    await userEvent.keyboard('{Backspace}');

    expect(onChange).toHaveBeenCalledWith(['javascript', 'svelte']);
  });

  it('does not add duplicate tags', async () => {
    const onChange = vi.fn();
    
    render(TagSelector, { 
      props: { tags: defaultTags, onChange } 
    });

    const input = screen.getByLabelText(/add new tag/i);
    await userEvent.type(input, 'javascript');
    await userEvent.keyboard('{Enter}');

    // onChange не должен быть вызван
    expect(onChange).not.toHaveBeenCalled();
    
    // Должна показаться ошибка
    expect(screen.getByRole('alert')).toHaveTextContent(/tag already exists/i);
  });

  it('validates tag length (max 30 chars)', async () => {
    const onChange = vi.fn();
    
    render(TagSelector, { 
      props: { tags: [], onChange } 
    });

    const input = screen.getByLabelText(/add new tag/i);
    // Вводим тег ровно 30 символов (граничное значение)
    const tag30 = 'a'.repeat(30);
    await userEvent.type(input, tag30);
    await userEvent.keyboard('{Enter}');

    // Тег с 30 символами должен быть добавлен
    expect(onChange).toHaveBeenCalledWith([tag30]);
    
    // Проверяем что maxlength работает - не даёт ввести больше 30
    expect(input).toHaveAttribute('maxlength', '30');
  });

  it('validates tag format (only alphanumeric, underscore, hyphen)', async () => {
    const onChange = vi.fn();
    
    render(TagSelector, { 
      props: { tags: [], onChange } 
    });

    const input = screen.getByLabelText(/add new tag/i);
    await userEvent.type(input, 'invalid tag!');
    await userEvent.keyboard('{Enter}');

    expect(onChange).not.toHaveBeenCalled();
    expect(screen.getByRole('alert')).toHaveTextContent(/can only contain letters/i);
  });

  it('enforces max tags limit', async () => {
    const onChange = vi.fn();
    
    render(TagSelector, { 
      props: { tags: ['tag1', 'tag2'], onChange, maxTags: 3 } 
    });

    // Добавляем третий тег
    const input = screen.getByLabelText(/add new tag/i);
    await userEvent.type(input, 'tag3');
    await userEvent.keyboard('{Enter}');

    expect(onChange).toHaveBeenCalledWith(['tag1', 'tag2', 'tag3']);

    // Проверяем что input скрылся после достижения лимита
    expect(screen.queryByLabelText(/add new tag/i)).not.toBeInTheDocument();
    
    // Проверяем что подсказка показывает лимит
    expect(screen.getByText(/3\/3 tags/i)).toBeInTheDocument();
  });

  it('hides input when max tags reached', () => {
    render(TagSelector, { 
      props: { tags: ['tag1', 'tag2', 'tag3'], maxTags: 3 } 
    });

    expect(screen.queryByLabelText(/add new tag/i)).not.toBeInTheDocument();
  });

  it('shows tag count hint', () => {
    render(TagSelector, { 
      props: { tags: defaultTags, maxTags: 10 } 
    });

    expect(screen.getByText(/3\/10 tags/i)).toBeInTheDocument();
  });

  it('clears error on new input', async () => {
    render(TagSelector, { 
      props: { tags: defaultTags } 
    });

    // Вызываем ошибку
    const input = screen.getByLabelText(/add new tag/i);
    await userEvent.type(input, 'javascript');
    await userEvent.keyboard('{Enter}');

    expect(screen.getByRole('alert')).toBeInTheDocument();

    // Начинаем вводить что-то новое
    await userEvent.type(input, 'x');

    // Ошибка должна исчезнуть
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('does not add empty tag', async () => {
    const onChange = vi.fn();
    
    render(TagSelector, { 
      props: { tags: [], onChange } 
    });

    const input = screen.getByLabelText(/add new tag/i);
    await userEvent.type(input, '   ');
    await userEvent.keyboard('{Enter}');

    expect(onChange).not.toHaveBeenCalled();
    expect(screen.getByRole('alert')).toHaveTextContent(/cannot be empty/i);
  });

  it('converts tags to lowercase', async () => {
    const onChange = vi.fn();
    
    render(TagSelector, { 
      props: { tags: [], onChange } 
    });

    const input = screen.getByLabelText(/add new tag/i);
    await userEvent.type(input, 'UPPERCASE');
    await userEvent.keyboard('{Enter}');

    expect(onChange).toHaveBeenCalledWith(['uppercase']);
  });

  it('has correct accessibility attributes', () => {
    render(TagSelector, { 
      props: { tags: defaultTags } 
    });

    // Каждый тег должен иметь кнопку удаления с aria-label
    defaultTags.forEach(tag => {
      expect(screen.getByLabelText(new RegExp(`remove tag ${tag}`, 'i'))).toBeInTheDocument();
    });

    // Поле ввода должно иметь aria-label
    expect(screen.getByLabelText(/add new tag/i)).toBeInTheDocument();
  });
});
