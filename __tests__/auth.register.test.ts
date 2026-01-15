import request from 'supertest';
import app from '../src/app';

describe('Auth API - Register', () => {
  const validUser = {
    name: 'Test User',
    email: 'test@example.com',
    password: 'Password@123',
  };

  it('should register a new user successfully', async () => {
    const res = await request(app)
      .post('/auth/register')
      .send(validUser);

    expect(res.status).toBe(201);

    expect(res.body).toHaveProperty('id');
    expect(res.body).toHaveProperty('name', validUser.name);
    expect(res.body).toHaveProperty('email', validUser.email);
    expect(res.body).not.toHaveProperty('password');
  });

  it('should fail if required fields are missing', async () => {
    const res = await request(app)
      .post('/auth/register')
      .send({
        email: validUser.email,
      });

    expect(res.status).toBe(400);
    expect(res.body).toHaveProperty('message');
  });

  it('should fail if email already exists', async () => {
    await request(app)
      .post('/auth/register')
      .send(validUser);

    const res = await request(app)
      .post('/auth/register')
      .send(validUser);

    expect(res.status).toBe(409);
    expect(res.body.message).toMatch(/email/i);
  });
});
