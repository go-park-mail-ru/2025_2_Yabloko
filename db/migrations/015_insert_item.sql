INSERT INTO
        item (id, name, description, card_img)
VALUES
        -- Наелся лосося
        (
                '9e1a4a7f-73a9-4b2d-9332-8aef8d3c11e1',
                'Ролл "Филадельфия"',
                'Нежный ролл с лососем, сливочным сыром и авокадо.',
                'naelsya_lososya_1.jpeg'
        ),
        (
                '1b0c9c25-3185-4c82-9154-79bb1ef5b53f',
                'Сет "Самурай"',
                '8 шт. ассорти: Филадельфия, Калифорния, Унаги, Спайси-тунец.',
                'naelsya_lososya_2.jpeg'
        ),
        -- Pizza Heart
        (
                '3b8dca61-2d4e-4f85-a75b-2ef29f7c13bb',
                'Пицца "Пепперони"',
                'Классика: томатный соус, моцарелла, острые колечки пепперони.',
                'pizza_heart_1.jpeg'
        ),
        (
                'f3c7b1ab-41ac-423a-b1cb-4d71b65bdf94',
                'Лимонад "Цитрусовый взрыв"',
                'Освежающий лимонад с лаймом, мятой и лёгкой газацией.',
                'pizza_heart_2.jpeg'
        ),
        -- SubJoy
        (
                'a4e8bcd0-44b1-40c4-b89a-8df76f36e2fa',
                'Сабвей "Турецкий"',
                'Курица-гриль, огурцы, томаты, лук, соус "Цезарь". 30 см.',
                'subjoy_1.jpeg'
        ),
        (
                'c5d7e2a9-6b1f-4de3-9b76-0c5e61d4f9b3',
                'Картофель по-деревенски',
                'Хрустящая картошка с прованскими травами и чесноком.',
                'subjoy_2.jpeg'
        ),
        -- Все шашлыки
        (
                'e8a0d27b-06e3-4bcb-90b9-2eac3b55e6de',
                'Шашлык из свинины',
                'Мягкая свинина маринованная в луке и специях. Подача с лавашом.',
                'vse_shashliki_1.jpeg'
        ),
        (
                '2f97bcdc-5e45-41cd-92a8-988bcf7eb3b2',
                'Пиво "Лагер" 0.5л',
                'Светлое нефильтрованное пиво местного производства.',
                'vse_shashliki_2.jpeg'
        );

INSERT INTO
        store_item (id, store_id, item_id, price)
VALUES
        (
                '7d1b5b40-0b94-4c4f-9e77-97aef6b0f1a3',
                'b2f0d6b3-65a2-4c2a-a32f-30a1b73f32e2',
                '9e1a4a7f-73a9-4b2d-9332-8aef8d3c11e1',
                420.00
        ),
        (
                'b6d9d82a-05c5-44b8-97c9-c45d1227a45f',
                'b2f0d6b3-65a2-4c2a-a32f-30a1b73f32e2',
                '1b0c9c25-3185-4c82-9154-79bb1ef5b53f',
                1250.00
        ),
        (
                '2a1c56b5-828e-47a9-8ef2-04b3e14e6c61',
                '9ac3b889-96df-4c93-a0b7-31f5b6a6e89c',
                '3b8dca61-2d4e-4f85-a75b-2ef29f7c13bb',
                590.00
        ),
        (
                '9f44b8b7-b88d-49a1-b36f-53d3b96d8d15',
                '9ac3b889-96df-4c93-a0b7-31f5b6a6e89c',
                'f3c7b1ab-41ac-423a-b1cb-4d71b65bdf94',
                180.00
        ),
        (
                '8e22f2c6-6b0b-4a5f-91cf-2c4e2c36c6b5',
                'c45a7b64-df32-4e84-b2cb-85a3b8e6b0fc',
                'a4e8bcd0-44b1-40c4-b89a-8df76f36e2fa',
                340.00
        ),
        (
                '1c77b2b0-5824-4d25-b8a9-1d5e734a8937',
                'c45a7b64-df32-4e84-b2cb-85a3b8e6b0fc',
                'c5d7e2a9-6b1f-4de3-9b76-0c5e61d4f9b3',
                150.00
        ),
        (
                'b3d2a7ea-8b8a-4822-912e-b2d14e8ac0f5',
                'd0c12a9f-2b2a-4e91-8e0a-13df58d9f8af',
                'e8a0d27b-06e3-4bcb-90b9-2eac3b55e6de',
                650.00
        ),
        (
                '6c5b7b7b-35f1-44b2-9d1a-6d6d8b5b7c7a',
                'd0c12a9f-2b2a-4e91-8e0a-13df58d9f8af',
                '2f97bcdc-5e45-41cd-92a8-988bcf7eb3b2',
                220.00
        ) ON CONFLICT (id) DO NOTHING;