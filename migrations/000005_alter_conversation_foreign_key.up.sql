ALTER TABLE conversation_participants
DROP CONSTRAINT conversation_participants_conversation_id_fkey;

ALTER TABLE conversation_participants
ADD CONSTRAINT conversation_participants_conversation_id_fkey
    FOREIGN KEY (conversation_id)
    REFERENCES conversations (id)
    ON DELETE CASCADE
    DEFERRABLE INITIALLY DEFERRED;
 