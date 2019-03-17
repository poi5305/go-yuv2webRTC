package app.robotmon.webrtcgo;

import android.Manifest;
import android.content.pm.PackageManager;
import android.graphics.Bitmap;
import android.graphics.BitmapFactory;
import android.graphics.ImageFormat;
import android.graphics.Rect;
import android.graphics.YuvImage;
import android.hardware.Camera;
import android.os.Build;
import android.support.v7.app.AppCompatActivity;
import android.os.Bundle;
import android.util.Log;
import android.util.TimingLogger;
import android.view.SurfaceHolder;
import android.view.SurfaceView;
import android.widget.Toast;

import java.io.ByteArrayOutputStream;
import java.io.IOException;

import gomobilelib.Gomobilelib;

public class MainActivity extends AppCompatActivity implements Camera.PreviewCallback, SurfaceHolder.Callback {

    private static final int PERMISSIONS_REQUEST = 1234;
    private static final int SRC_FRAME_WIDTH = 640;
    private static final int SRC_FRAME_HEIGHT = 480;

    private Camera mCamera;
    private Camera.Parameters mParams;
    private SurfaceView mSurfaceView;
    private SurfaceHolder mSurfaceHolder;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        mSurfaceView = findViewById(R.id.preview);
        mSurfaceHolder = mSurfaceView.getHolder();
        mSurfaceHolder.setFixedSize(SRC_FRAME_WIDTH, SRC_FRAME_HEIGHT);
        mSurfaceHolder.addCallback(this);
        mSurfaceHolder.setType(SurfaceHolder.SURFACE_TYPE_PUSH_BUFFERS);

        // Call go to init webRTC and serve a web
        Gomobilelib.initWebRTC(SRC_FRAME_WIDTH, SRC_FRAME_HEIGHT);

        openCamera();
    }

    private void openCamera() {
        if (!hasPermission()) {
            requestPermission();
            return;
        }
        if (mCamera != null) {
            return;
        }
        mCamera = Camera.open(Camera.CameraInfo.CAMERA_FACING_BACK);
        mParams = mCamera.getParameters();
        mParams.setPreviewSize(SRC_FRAME_WIDTH, SRC_FRAME_HEIGHT);
        mParams.setPreviewFormat(ImageFormat.NV21);
        mCamera.setParameters(mParams); // setting camera parameters
        mCamera.setPreviewCallback(this);
        try {
            mCamera.setPreviewDisplay(mSurfaceHolder);
            mCamera.startPreview();
        } catch (IOException ioe) {
            ioe.printStackTrace();
        }
    }

    private void releaseCamera() {
        if (mCamera != null) {
            try {
                mCamera.setPreviewCallback(null);
                mCamera.stopPreview();
                mCamera.release();
            } catch (Exception e) {
                e.printStackTrace();
            }
            mCamera = null;
        }
    }

    private boolean hasPermission() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            return checkSelfPermission(Manifest.permission.CAMERA) == PackageManager.PERMISSION_GRANTED;
        }
        return true;
    }

    private void requestPermission() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            requestPermissions(new String[]{Manifest.permission.CAMERA}, PERMISSIONS_REQUEST);
        }
    }

    @Override
    public void onResume() {
        super.onResume();
        openCamera();
    }

    @Override
    protected void onPause() {
        super.onPause();
        releaseCamera();
    }

    @Override
    public void onPreviewFrame(final byte[] data, Camera camera) {
        // Call go to prepare webRTC streaming vpx
        Gomobilelib.onPreviewFrame(data);
    }

    @Override
    public void onRequestPermissionsResult(
            final int requestCode, final String[] permissions, final int[] grantResults) {
        if (requestCode == PERMISSIONS_REQUEST) {
            if (grantResults.length > 0 && grantResults[0] == PackageManager.PERMISSION_GRANTED) {
                openCamera();
            } else {
                requestPermission();
            }
        }
    }

    @Override
    public void surfaceCreated(SurfaceHolder holder) {
        openCamera(); // open camera
    }

    @Override
    public void surfaceChanged(SurfaceHolder holder, int format, int width, int height) {
        try {
            mCamera.setPreviewDisplay(holder);
            mCamera.startPreview();
        } catch (IOException ioe) {
            ioe.printStackTrace();
        }
    }

    @Override
    public void surfaceDestroyed(SurfaceHolder holder) {
        releaseCamera();
    }

}
