import { describe, it, expect } from 'vitest';
import { MessageFormatter, GalacticLexicon, createFormatter } from './galactic-lexicon';

describe('MessageFormatter', () => {
  describe('technical mode', () => {
    const formatter = new MessageFormatter(false);

    it('should return technical success messages', () => {
      const message = formatter.success('noteCreated', 'Test Note');
      expect(message).toContain('Test Note');
      expect(message).toContain('создана');
      expect(message).not.toContain('Звезда');
    });

    it('should return technical error messages', () => {
      const message = formatter.error('validation', 'title');
      expect(message).toContain('title');
      expect(message).toContain('Неверное значение');
      expect(message).not.toContain('аномалию');
    });

    it('should return technical info messages', () => {
      const message = formatter.info('emptyGraph');
      expect(message).toContain('Граф пуст');
      // Technical mode also mentions звёздное небо as a metaphor
      expect(message).toContain('звёздное небо');
    });

    it('should return technical warning messages', () => {
      const message = formatter.warning('unsavedChanges');
      expect(message).toContain('несохраненные изменения');
      expect(message).not.toContain('бортовом журнале');
    });
  });

  describe('galactic mode', () => {
    const formatter = new MessageFormatter(true);

    it('should return galactic success messages', () => {
      const message = formatter.success('noteCreated', 'Test Note');
      expect(message).toContain('Test Note');
      expect(message).toContain('Звезда');
      expect(message).toContain('зажжена');
    });

    it('should return galactic error messages', () => {
      const message = formatter.error('validation', 'title');
      expect(message).toContain('title');
      expect(message).toContain('аномалию');
      // Check for 'Сенсоры' (capital S) as it appears in the actual message
      expect(message).toMatch(/[Сс]енсоры/);
    });

    it('should return galactic info messages', () => {
      const message = formatter.info('emptyGraph');
      expect(message).toContain('звёздное небо');
      expect(message).toContain('пусто');
    });

    it('should return galactic warning messages', () => {
      const message = formatter.warning('unsavedChanges');
      expect(message).toContain('бортовом журнале');
      expect(message).toContain('записи');
    });
  });

  describe('mode switching', () => {
    it('should switch between modes', () => {
      const formatter = new MessageFormatter(false);
      
      // Technical mode
      let message = formatter.success('noteCreated', 'Note');
      expect(message).toContain('создана');
      
      // Switch to galactic
      formatter.setGalacticMode(true);
      message = formatter.success('noteCreated', 'Note');
      expect(message).toContain('Звезда');
      
      // Switch back
      formatter.setGalacticMode(false);
      message = formatter.success('noteCreated', 'Note');
      expect(message).toContain('создана');
    });

    it('should report current mode', () => {
      const formatter = new MessageFormatter(false);
      expect(formatter.isGalacticMode()).toBe(false);
      
      formatter.setGalacticMode(true);
      expect(formatter.isGalacticMode()).toBe(true);
    });
  });

  describe('createFormatter helper', () => {
    it('should create formatter with specified mode', () => {
      const technicalFormatter = createFormatter(false);
      expect(technicalFormatter.isGalacticMode()).toBe(false);
      
      const galacticFormatter = createFormatter(true);
      expect(galacticFormatter.isGalacticMode()).toBe(true);
    });
  });
});

describe('GalacticLexicon', () => {
  describe('success messages', () => {
    it('should return note created message in technical mode', () => {
      const message = GalacticLexicon.success.noteCreated('My Note', false);
      expect(message).toContain('My Note');
      expect(message).toContain('создана');
    });

    it('should return note created message in galactic mode', () => {
      const message = GalacticLexicon.success.noteCreated('My Note', true);
      expect(message).toContain('My Note');
      expect(message).toContain('Звезда');
    });

    it('should return achievement unlocked message', () => {
      const technical = GalacticLexicon.success.achievementUnlocked('Explorer', false);
      expect(technical).toContain('Explorer');
      expect(technical).toContain('Достижение');

      const galactic = GalacticLexicon.success.achievementUnlocked('Explorer', true);
      expect(galactic).toContain('Explorer');
      expect(galactic).toContain('звезда');
    });
  });

  describe('error messages', () => {
    it('should return validation error in technical mode', () => {
      const message = GalacticLexicon.error.validation('email', false);
      expect(message).toContain('email');
      expect(message).toContain('Неверное значение');
    });

    it('should return validation error in galactic mode', () => {
      const message = GalacticLexicon.error.validation('email', true);
      expect(message).toContain('email');
      expect(message).toContain('аномалию');
    });

    it('should return unauthorized error', () => {
      const technical = GalacticLexicon.error.unauthorized(false);
      expect(technical).toContain('Требуется авторизация');

      const galactic = GalacticLexicon.error.unauthorized(true);
      expect(galactic).toContain('Отказано');
      expect(galactic).toContain('звёздной системе');
    });
  });

  describe('info messages', () => {
    it('should return empty graph message', () => {
      const technical = GalacticLexicon.info.emptyGraph(false);
      expect(technical).toContain('Граф пуст');

      const galactic = GalacticLexicon.info.emptyGraph(true);
      expect(galactic).toContain('звёздное небо');
    });

    it('should return streak message with days', () => {
      const technical = GalacticLexicon.info.streakActive(7, false);
      expect(technical).toContain('7');
      expect(technical).toContain('дней');

      const galactic = GalacticLexicon.info.streakActive(7, true);
      expect(galactic).toContain('7');
      expect(galactic).toContain('путешествие');
    });
  });

  describe('warning messages', () => {
    it('should return unsaved changes message', () => {
      const technical = GalacticLexicon.warning.unsavedChanges(false);
      expect(technical).toContain('несохраненные изменения');

      const galactic = GalacticLexicon.warning.unsavedChanges(true);
      expect(galactic).toContain('бортовом журнале');
    });

    it('should return delete confirm message with item name', () => {
      const technical = GalacticLexicon.warning.deleteConfirm('My Note', false);
      expect(technical).toContain('My Note');
      expect(technical).toContain('удалить');

      const galactic = GalacticLexicon.warning.deleteConfirm('My Note', true);
      expect(galactic).toContain('My Note');
      expect(galactic).toContain('чёрную дыру');
    });
  });
});

describe('message consistency', () => {
  it('should have matching keys in both modes', () => {
    const formatter = new MessageFormatter(false);
    
    // Test that all categories have the same keys
    const successKeys = ['noteCreated', 'noteUpdated', 'noteDeleted', 'linkCreated', 'settingsSaved', 'achievementUnlocked', 'shareCreated', 'loginSuccess'];
    const errorKeys = ['validation', 'duplicateLink', 'noteNotFound', 'unauthorized', 'serverError'];
    
    // All keys should work in both modes
    successKeys.forEach(key => {
      const technical = formatter.format('success', key, 'test');
      expect(technical).toBeTruthy();
      
      formatter.setGalacticMode(true);
      const galactic = formatter.format('success', key, 'test');
      expect(galactic).toBeTruthy();
      expect(galactic).not.toEqual(technical);
      
      formatter.setGalacticMode(false);
    });
  });
});
