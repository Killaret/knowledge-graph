import { Given, When, Then } from "@cucumber/cucumber";
import { expect, type Browser } from "@playwright/test";
import type { ITestWorld } from "../support/world";
import { loginOrCreateTestUser } from "../../helpers/auth";

const getFrontendUrl = (): string =>
  (process.env.FRONTEND_URL || "http://localhost:5173").replace(/\/$/, "");

Given("the test user exists", async function (this: ITestWorld) {
  await loginOrCreateTestUser(this.request);
});

Given("I start an anonymous session", async function (this: ITestWorld) {
  const browser = this.context?.browser() as Browser | undefined;
  if (!browser) {
    throw new Error("Cannot access browser from current context");
  }
  await this.context?.close();
  this.context = await browser.newContext({
    viewport: { width: 1280, height: 720 },
  });
  this.page = await this.context.newPage();
  this.request = this.context.request;
});

Given("I am on the login page", async function (this: ITestWorld) {
  await this.page.goto(`${getFrontendUrl()}/auth/login`);
  await this.page.waitForLoadState("domcontentloaded");
  await this.page.waitForTimeout(500);
});

Given("I am on the registration page", async function (this: ITestWorld) {
  await this.page.goto(`${getFrontendUrl()}/auth/register`);
  await this.page.waitForLoadState("domcontentloaded");
  await this.page.waitForTimeout(500);
});

When(
  "I enter {string} and {string}",
  async function (this: ITestWorld, login: string, password: string) {
    await this.page.locator("input#login").fill(login);
    await this.page.locator("input#password").fill(password);
  }
);

When("I click the sign in button", async function (this: ITestWorld) {
  await this.page.locator('button[type="submit"]').click();
});

When(
  "I fill in the registration form with a unique login and password {string}",
  async function (this: ITestWorld, password: string) {
    const uniqueLogin = `bdduser-${Date.now()}`;
    const email = `${uniqueLogin}@example.com`;
    await this.page.locator("input#login").fill(uniqueLogin);
    await this.page.locator("input#email").fill(email);
    await this.page.locator("input#password").fill(password);
    await this.page.locator("input#confirm-password").fill(password);
  }
);

When("I click the register button", async function (this: ITestWorld) {
  await this.page.locator('button[type="submit"]').click();
});

Then("I should be redirected to the main page", async function (this: ITestWorld) {
  await this.page.waitForURL(/\/$/, { timeout: 10000 });
  await expect(this.page).toHaveURL(/\/$/);
});

Then("I should be redirected to {string}", async function (this: ITestWorld, path: string) {
  const escaped = path.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const regex = new RegExp(escaped);
  await this.page.waitForURL(regex, { timeout: 10000 });
  await expect(this.page).toHaveURL(regex);
});

Then("I should see an authentication error", async function (this: ITestWorld) {
  const error = this.page.locator('[role="alert"], .error-container').first();
  await expect(error).toBeVisible({ timeout: 5000 });
});
