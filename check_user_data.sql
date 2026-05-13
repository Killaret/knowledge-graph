-- Проверка пользователя
SELECT 
    id,
    username,
    email,
    lastname,
    firstname,
    userpassword,
    activated,
    email_verified,
    phone_number_confirmed,
    created_at,
    updated_at
FROM public.users 
WHERE username = 'cwwc';

-- Проверка групп пользователя
SELECT 
    uig.id,
    uig.user_id,
    uig.group_id,
    g.name as group_name,
    g.description
FROM public.users_in_groups uig
JOIN public.groups g ON uig.group_id = g.id
WHERE uig.user_id = '7b87df8d-8a12-476c-8bdc-df01eef9fe0c'::uuid;

-- Проверка всех доступных групп
SELECT 
    id,
    name,
    description,
    created_at
FROM public.groups 
ORDER BY name;
