BEGIN;

UPDATE articles
SET image_url = CASE
        WHEN image_url = '/images/app_logo.png' THEN '/images/articles/change-name-district-office.webp'
        ELSE image_url
    END,
    excerpt = 'เช็กลิสต์เอกสาร ขั้นตอน ค่าธรรมเนียม และหลักตั้งชื่อตัว–ชื่อสกุล พร้อมวิธีใช้ชื่อดี 4 มิติช่วยคัดชื่อก่อนไปอำเภอหรือสำนักงานเขต',
    updated_at = NOW()
WHERE slug = 'เปลี่ยนชื่อที่อำเภอต้องเตรียมอะไรบ้าง';

COMMIT;
