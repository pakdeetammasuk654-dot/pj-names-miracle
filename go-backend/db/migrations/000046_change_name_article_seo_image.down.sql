BEGIN;

UPDATE articles
SET image_url = CASE
        WHEN image_url = '/images/articles/change-name-district-office.webp' THEN '/images/app_logo.png'
        ELSE image_url
    END,
    excerpt = 'เช็กลิสต์เอกสาร ขั้นตอน ค่าธรรมเนียม และหลักตั้งชื่อตัว–ชื่อสกุลให้ผ่านการตรวจของนายทะเบียน พร้อมวิธีใช้แอปชื่อดี 4 มิติช่วยคัดชื่อก่อนเดินทางไปอำเภอหรือสำนักงานเขต',
    updated_at = NOW()
WHERE slug = 'เปลี่ยนชื่อที่อำเภอต้องเตรียมอะไรบ้าง';

COMMIT;
