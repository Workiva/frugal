package com.workiva.frugal.transport;

/**
 * This class is used for carrying pointer to Amazon S3 objects which contain
 * message payloads. For a large-payload messages, an instance of this class
 * will be serialized to JSON and sent through Amazon SQS.
 *
 */
public class MessageS3Pointer {
	private String s3BucketName;
	private String s3Key;

	public MessageS3Pointer() {
	}

	public MessageS3Pointer(String s3BucketName, String s3Key) {
		this.s3BucketName = s3BucketName;
		this.s3Key = s3Key;
	}

	public String getS3BucketName() {
		return s3BucketName;
	}

	public void setS3BucketName(String s3BucketName) {
		this.s3BucketName = s3BucketName;
	}

	public String getS3Key() {
		return s3Key;
	}

	public void setS3Key(String s3Key) {
		this.s3Key = s3Key;
	}

}
